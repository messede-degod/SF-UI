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
	"strings"
	"sync/atomic"
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
	StartXpraCommand         string `yaml:"start_xpra_command"`          // Command used to start xpra
	StartVNCCommand          string `yaml:"start_vnc_command"`           // Command used to start VNC
	StartFileBrowserCommand  string `yaml:"start_filebrowser_command"`   // Command used to start filebrowser
	VNCPort                  uint16 `yaml:"vnc_port"`
	FileBrowserPort          uint16 `yaml:"filebrowser_port"`

	CompiledClientConfig   []byte   // Ui related config that has to be sent to client
	SfEndpoints            []string `yaml:"sf_endpoints"`               // Sf Endpoints To Use
	SfUIOrigin             string   `yaml:"sf_ui_origin"`               // Where SFUI is deployed, for CSRF prevention, ex: https://web.segfault.net
	UseXForwardedForHeader bool     `yaml:"use_x_forwarded_for_header"` // Use the X-Forwared-For HTTP header, usefull when behind a reverse proxy
	DisableOriginCheck     bool     `yaml:"disable_origin_check"`       // Disable Origin Checking
	DisableDesktop         bool     `yaml:"disable_desktop"`            // Disable websocket based GUI desktop access

	ClientInactivityTimeout int                 `yaml:"client_inactivity_timeout"` // Minutes after which the clients master SSH connection is killed
	ValidSecret             func(s string) bool // Secret Validator
	EndpointSelector        *atomic.Int32       // Helps select a endpoint in RR fashion
	NoEndpoints             int32               // No of available endpoints

	SegfaultSSHUsername string `yaml:"segfault_ssh_username"`
	SegfaultSSHPassword string `yaml:"segfault_ssh_password"`
	SegfaultUseSSHKey   bool   `yaml:"segfault_use_ssh_key"`  // whether to use a ssh key
	SegfaultSSHKeyPath  string `yaml:"segfault_ssh_key_path"` // absolute path to the ssh key
}

var buildTime string
var buildHash string
var SfuiVersion string = "0.2.0"

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
		sfui.handleLogin(w, r)
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

func (sfui *SfUI) handleLogin(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		loginReq := TermRequest{}
		if json.Unmarshal(data, &loginReq) == nil {
			loginReq.ClientIp = sfui.getClientAddr(r)
			if loginReq.NewInstance {
				secret := sfui.getEndpointNameRR() + "-"
				secret += sfui.generateSecret(&loginReq)

				w.WriteHeader(http.StatusOK)
				termRes := TermResponse{
					Status: "OK",
					Secret: secret,
				}
				response, _ := json.Marshal(termRes)
				w.Write(response)
				return
			}

			if sfui.ValidSecret(loginReq.Secret) {
				client, cerr := sfui.GetClient(loginReq.Secret)
				isDuplicate := false
				if cerr == nil {
					// 1 active and non matching tab ids - Duplicate
					// 2 active and matching tab ids - Non Duplicate
					if client.TabId != nil && client.ClientActive != nil {
						winIdMatches := (*client.TabId == loginReq.TabId)

						if client.ClientActive.Load() && !winIdMatches {
							isDuplicate = true
						}

						// 3 inactive and matching tab ids - Non Duplicate
						// 4 inactive and non matching tab ids - Non Duplicate, set new tab id
						if !client.ClientActive.Load() && !winIdMatches {
							client.SetTabId(loginReq.TabId)
						}
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
					sfui.RemoveClient(&client)
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

// Split secret into endpoint name and actual-secret
// return the endpoint FQDN based on the name (ex: 8lgm -> return 8lgm.segfault.net)
// defaults to first available endpoint FQDN if name is not found.
func (sfui *SfUI) getEndpointAndSecret(secret string) (EndpointAddress string, ActualSecret string) {
	secretParts := strings.Split(secret, "-") // secret is in the form  "endpointname-randomsecretXXXXX"
	if len(secretParts) > 1 {
		endpointName := secretParts[0]

		for _, address := range sfui.SfEndpoints {
			if strings.Contains(address, endpointName) {
				return address, secretParts[1]
			}
		}
	}

	return sfui.SfEndpoints[0], secret
}

func (sfui *SfUI) getEndpointNameRR() string {
	selected := sfui.EndpointSelector.Load()
	if selected > sfui.NoEndpoints-1 {
		sfui.EndpointSelector.Store(0)
		selected = 0
	}
	sfui.EndpointSelector.Add(1)

	eparts := strings.Split(sfui.SfEndpoints[selected], ".")
	if len(eparts) > 0 {
		return eparts[0]
	}

	return ""
}
