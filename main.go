package main

import (
	"embed"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type SfUI struct {
	MaxWsTerminals    int    `yaml:"max_ws_terminals"`    // Max terminals that can be allocated per client
	ServerBindAddress string `yaml:"server_bind_address"` // Address to which the current app binds
	XpraWSAddress     string `yaml:"xpra_ws_address"`     // Address at which the xpra ws server is listening
	Debug             bool   `yaml:"debug"`               // Print debug information

	MasterSSHCommand         string `yaml:"master_ssh_command"`          // Command used to setup the SSH Master Socket
	TearDownMasterSSHCommand string `yaml:"teardown_master_ssh_command"` // Command used to teardown the SSH Master Socket.
	SlaveSSHCommand          string `yaml:"slave_ssh_command"`           // Command used to start a SSH shell using the master socket
	GUIBridgeCommand         string `yaml:"gui_bridge_command"`          // Command used to setup a GUI port forward using the master socket

	CompiledClientConfig []byte // Ui related config that has to be sent to client
	SfEndpoint           string `yaml:"sf_endpoint"`          // Current Sf Endpoints Name
	SfUIOrigin           string `yaml:"sf_ui_origin"`         // Where SFUI is deployed, for CSRF prevention, ex: https://web.segfault.net
	DisableOriginCheck   bool   `yaml:"disable_origin_check"` // Disable Origin Checking
	DisableDesktop       bool   `yaml:"disable_desktop"`      // Disable websocket based GUI desktop access
	// Directory where SSH sockets are stored,
	// Diretcory Structure:
	// 		WorkDir/
	//			|-sfui/		(created by sfui- container for client dirs)
	//				|-perClientUniqDir/ (a unique string derived from secret)
	//						- gui.sock (ssh -L ./gui.sock:127.0.0.1:2000 root@segfault.net)
	WorkDirectory string `yaml:"work_directory"`
}

var buildTime string
var SfuiVersion string = "0.1.1"

//go:embed ui/dist/sf-ui
var staticfiles embed.FS

func main() {
	sfui := ReadConfig()
	gerr := sfui.cleanWorkDir()
	if gerr != nil {
		log.Fatal(gerr)
	}

	log.Printf("SFUI [Version : %s] [Built on : %s]\n", SfuiVersion, buildTime)
	log.Printf("Listening on http://%s ....\n", sfui.ServerBindAddress)
	http.ListenAndServe(sfui.ServerBindAddress, http.HandlerFunc(sfui.requestHandler))
}

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
	case "/config":
		sfui.handleUIConfig(w, r)
		w.Header().Add("Content-Type", "application/json")
	case "/ws":
		sfui.handleTerminalWs(w, r)
	case "/xpraws":
		if !sfui.DisableDesktop {
			sfui.handleDesktopWS(w, r)
		}
	default:
		handleUIRequest(w, r)
	}
}

func (sfui *SfUI) handleSecret(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		termReq := TermRequest{}
		if json.Unmarshal(data, &termReq) == nil {
			termReq.ClientIp = r.RemoteAddr
			if termReq.NewInstance {
				secret, err := sfui.generateSecret(&termReq)
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

			if sfui.secretValid(&termReq) == nil {
				w.WriteHeader(http.StatusOK)
				termRes := TermResponse{
					Status: "OK",
					// SfEndpoint: sfui.SfEndpoint,
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
