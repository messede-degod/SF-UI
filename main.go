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
	ShellCommand      string `yaml:"shell_command"`       // Command to run on the pty
	// Additional arguments are inserted into to ShellCommand
	// Expects ShellCommand to have format specifiers.
	// "%[1]s" will be replaced by the client secret.
	// "%[2]s" will be replaced by the clients ip address.
	// 	Example:  mycmd -secret %[1]s -client-ip %[2]s
	// 	Result: mycmd -secret XXXXXX -client-ip 7.7.7.7
	// This is usefull for global deployment of SFUI and integration
	// with Segfault Core.
	// If false, user is redirected to SFUI dashboard without any authentication
	AddSfUIArgs          bool   `yaml:"add_sf_ui_args"`
	CompiledClientConfig []byte // Ui related config that has to be sent to client
	SfEndpoint           string `yaml:"sf_endpoint"`          // Current Sf Endpoints Name
	SfUIOrigin           string `yaml:"sf_ui_origin"`         // Where SFUI is deployed, for CSRF prevention, ex: https://web.segfault.net
	DisableOriginCheck   bool   `yaml:"disable_origin_check"` // Disable Origin Checking
}

var buildTime string

//go:embed ui/dist/sf-ui
var staticfiles embed.FS

func main() {
	sfui := ReadConfig()

	log.Printf("SFUI [Version : %s] [Built on : %s]\n", "0.1", buildTime)
	log.Printf("Listening on http://%s ....\n", sfui.ServerBindAddress)
	http.ListenAndServe(sfui.ServerBindAddress, http.HandlerFunc(sfui.requestHandler))
}

func (sfui *SfUI) requestHandler(w http.ResponseWriter, r *http.Request) {
	if sfui.Debug {
		log.Println(r.RemoteAddr, " ", r.URL, " ", r.UserAgent())
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
		sfui.handleDesktopWS(w, r)
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
