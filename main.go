package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
)

type SfUI struct {
	MaxWsTerminals       int    `yaml:"max_ws_terminals"`        // Max terminals that can be allocated per client
	MaxSharedDesktopConn int    `yaml:"max_shared_desktop_conn"` // Max no of clients that can connect to a shared desktop
	WSPingInterval       int    `yaml:"ws_ping_interval"`        // Intervals at which the client pings the terminals WS connection
	WSTimeout            int    `yaml:"ws_timeout"`              // Timeout (in minutes) applied to terminal and desktop ws connections
	ServerBindAddress    string `yaml:"server_bind_address"`     // Address to which the current app binds
	Debug                bool   `yaml:"debug"`                   // Print debug information

	MasterSSHCommand         string `yaml:"master_ssh_command"`          // Command used to setup the SSH Master Socket
	TearDownMasterSSHCommand string `yaml:"teardown_master_ssh_command"` // Command used to teardown the SSH Master Socket.
	SlaveSSHCommand          string `yaml:"slave_ssh_command"`           // Command used to start a SSH shell using the master socket
	GUIBridgeCommand         string `yaml:"gui_bridge_command"`          // Command used to setup a GUI port forward using the master socket
	StartXpraCommand         string `yaml:"start_xpra_command"`          // Command used to start xpra
	StartVNCCommand          string `yaml:"start_vnc_command"`           // Command used to start VNC
	StartFileBrowserCommand  string `yaml:"start_filebrowser_command"`   // Command used to start filebrowser

	CompiledClientConfig   []byte // Ui related config that has to be sent to client
	SfEndpoint             string `yaml:"sf_endpoint"`                // Current Sf Endpoints Name
	SfUIOrigin             string `yaml:"sf_ui_origin"`               // Where SFUI is deployed, for CSRF prevention, ex: https://web.segfault.net
	UseXForwardedForHeader bool   `yaml:"use_x_forwarded_for_header"` // Use the X-Forwared-For HTTP header, usefull when behind a reverse proxy
	DisableOriginCheck     bool   `yaml:"disable_origin_check"`       // Disable Origin Checking
	DisableDesktop         bool   `yaml:"disable_desktop"`            // Disable websocket based GUI desktop access
	// Directory where SSH sockets are stored,
	// Diretcory Structure:
	// 		WorkDir/
	//			|-sfui/		(created by sfui- container for client dirs)
	//				|-perClientUniqDir/ (a unique string derived from secret)
	//						- gui.sock (ssh -L ./gui.sock:127.0.0.1:2000 root@segfault.net)
	WorkDirectory           string              `yaml:"work_directory"`
	ClientInactivityTimeout int                 `yaml:"client_inactivity_timeout"` // Minutes after which the clients master SSH connection is killed
	ValidSecret             func(s string) bool // Secret Validator
}

var buildTime string
var buildHash string
var SfuiVersion string = "0.1.1"

//go:embed ui/dist/sf-ui
var staticfiles embed.FS

func main() {
	if ActionInvoked := handleCmdLineFlags(); ActionInvoked {
		return
	}

	sfui := ReadConfig()
	log.Printf("SFUI [Version : %s] [Built on : %s]\n", SfuiVersion, buildTime)

	rlErr := obtainRunLock()
	if rlErr != nil {
		log.Println(rlErr)
		return
	}
	// release runLock in cleanUp()

	gerr := sfui.cleanWorkDir()
	if gerr != nil {
		log.Fatal(gerr)
	}

	sfui.handleSignals()

	log.Printf("Listening on http://%s ....\n", sfui.ServerBindAddress)
	http.ListenAndServe(sfui.ServerBindAddress, http.HandlerFunc(sfui.requestHandler))
}

func (sfui *SfUI) handleSignals() {
	sigs := make(chan os.Signal, 1)
	// catch all signals
	signal.Notify(sigs)

	go func() {
		for sig := range sigs {
			switch sig {
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				fallthrough
			case syscall.SIGHUP:
				sfui.cleanUp()
				os.Exit(0)
			}
		}
	}()
}

func handleCmdLineFlags() (ActionInvoked bool) {
	// Handle CmdLine Flags
	var install bool
	var uninstall bool

	flag.BoolVar(&install, "install", false, "install SFUI")
	flag.BoolVar(&uninstall, "uninstall", false, "uninstall SFUI")
	flag.Parse()

	if install {
		ierr := InstallService()
		if ierr != nil {
			log.Println(ierr.Error())
		}
		ActionInvoked = true
	}

	if uninstall {
		uierr := UnInstallService()
		if uierr != nil {
			log.Println(uierr.Error())
		}
		ActionInvoked = true
	}

	return ActionInvoked
}

func (sfui *SfUI) cleanUp() {
	sfui.DisableClientAccess()
	log.Println("Disconnecting all clients...")
	sfui.RemoveAllClients()
	releaseRunLock()
}

var isFbPath = regexp.MustCompile(`(?m)^/filebrowser.*`).MatchString

func (sfui *SfUI) requestHandler(w http.ResponseWriter, r *http.Request) {
	if sfui.Debug {
		// log.Println(r.RemoteAddr, " ", r.URL, " ", r.UserAgent())
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "*")
		w.Header().Add("Access-Control-Allow-Headers", "*")
	}
	switch r.URL.Path {
	case "/secret":
		sfui.handleSecret(w, r)
		w.Header().Add("Content-Type", "application/json")
	case "/logout":
		sfui.handleLogout(w, r)
	case "/config":
		sfui.handleUIConfig(w, r)
		w.Header().Add("Content-Type", "application/json")
	case "/ws":
		sfui.handleTerminalWs(w, r)
	case "/desktopws":
		if !sfui.DisableDesktop {
			sfui.handleDesktopWS(w, r)
		}
	case "/sharedDesktopWs":
		if !sfui.DisableDesktop {
			sfui.handleSharedDesktopWS(w, r)
		}
	case "/filebrowser":
		sfui.handleSetupFileBrowser(w, r)
	case "/desktop/share":
		sfui.handleSetupDesktopSharing(w, r)
		w.Header().Add("Content-Type", "application/json")
	default:
		// /filebrowser/*
		if isFbPath(r.URL.Path) {
			sfui.handleFileBrowser(w, r)
			return
		}
		handleUIRequest(w, r)
	}
}

func (sfui *SfUI) handleSecret(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		loginReq := TermRequest{}
		if json.Unmarshal(data, &loginReq) == nil {
			loginReq.ClientIp = sfui.getClientAddr(r)
			if loginReq.NewInstance {
				secret, err := sfui.generateSecret(&loginReq)
				if err == nil {
					w.WriteHeader(http.StatusOK)
					termRes := TermResponse{
						Status: "OK",
						Secret: secret,
					}
					response, _ := json.Marshal(termRes)
					w.Write(response)
					return
				}
			}

			if sfui.ValidSecret(loginReq.Secret) {
				client, cerr := sfui.GetClient(loginReq.Secret)
				isDuplicate := false
				if cerr == nil {
					// 1 active and non matching tab ids - Duplicate
					// 2 active and matching tab ids - Non Duplicate
					winIdMatches := (client.TabId == loginReq.TabId)

					if client.ClientActive && !winIdMatches {
						isDuplicate = true
					}

					// 3 inactive and matching tab ids - Non Duplicate
					// 4 inactive and non matching tab ids - Non Duplicate, set new tab id
					if !client.ClientActive && !winIdMatches {
						client.SetTabId(loginReq.TabId)
					}
				} else {
					// start a new client
					go func() {
						client, cerr := sfui.GetExistingClientOrMakeNew(loginReq.Secret, loginReq.ClientIp)
						if cerr == nil {
							client.SetTabId(loginReq.TabId)
						}
					}()
				}

				w.WriteHeader(http.StatusOK)
				termRes := TermResponse{
					Status:      "OK",
					IsDuplicate: isDuplicate,
				}
				response, _ := json.Marshal(termRes)
				w.Write(response)
				return
			}
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"Internal Server Error"}`))
}

func (sfui *SfUI) handleLogout(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		logoutReq := TermRequest{}
		if json.Unmarshal(data, &logoutReq) == nil {
			if sfui.ValidSecret(logoutReq.Secret) {
				// Remove the client connection
				client, err := sfui.GetClient(logoutReq.Secret)
				if err == nil { // Client exists
					if client.mu != nil {
						client.mu.Lock()
						defer client.mu.Unlock()
						fclient, ok := clients[client.ClientId]
						if !ok {
							w.WriteHeader(http.StatusUnavailableForLegalReasons)
							w.Write([]byte(`{"status":"client not present"}`))
							return
						}
						if !fclient.ClientActive { // Make sure RemoveClientIfInactive doesnt try to remove the client once more
							close(fclient.ClientConn) // Marking client as active to prevent RemoveClientIfInactive from running
						}
						sfui.RemoveClient(&client)
						// No need to unlock client since its now deleted
					}
				}

				w.WriteHeader(http.StatusOK)
				termRes := TermResponse{
					Status: "OK",
				}
				response, _ := json.Marshal(termRes)
				w.Write(response)
				return
			}
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"Internal Server Error"}`))
}
