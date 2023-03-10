package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/koding/websocketproxy"
	"golang.org/x/net/websocket"
)

type SfUI struct {
	MaxWsTerminals    int    // Max terminals that can be allocated per client
	ServerBindAddress string // Address to which the current app binds
	XpraWSAddress     string // Address at which the xpra ws server is listening
	Debug             bool   // Print debug information
	ShellCommand      string // Command to run on the pty
	// Two additional arguments are added to ShellCommand
	// 	Example:  somecmd  SECRET=abc REMOTE_ADDR=1.1.1.1
	// This is usefull for global deployment of SFUI and integration
	// with Segfault Core.
	// If false, user is redirected to SFUI dashboard without any authentication
	AddSfUIArgs          bool
	CompiledClientConfig []byte // Ui related onfig that has to be sent to client
	// SfEndpoint           string  // Current Sf Endpoints Name
	SfUIOrigin         string // Where SFUI is deployed, for CSRF prevention, ex: https://web.segfault.net
	DisableOriginCheck bool   // Disable Origin Checking

}

type TermRequest struct {
	Secret      string `json:"secret"`
	NewInstance bool   `json:"new_instance"`
	ClientIp    string
	// SfEndpoint string `json:"sf_endpoint"`
}

type TermResponse struct {
	Status string `json:"status"`
	Secret string `json:"secret,omitempty"`
	// SfEndpoint string `json:"sf_endpoint"`
}

// First byte read from Terminal.Pty is matched with
// the following constants, to determine the type of data
// a client is sending
const (
	SFUI_CMD_RESIZE = '1'
	SFUI_CMD_PAUSE  = '2'
	SFUI_CMD_RESUME = '3'
)

type Terminal struct {
	ClientSecret string
	ClientIp     string
	TermConfig   *TermConfig
	WSConn       *websocket.Conn
	Pty          *os.File
}

type TermConfig struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

var buildTime string

//go:embed ui/dist/sf-ui
var staticfiles embed.FS

func main() {
	sfui := SfUI{
		MaxWsTerminals:     10,
		ServerBindAddress:  "127.0.0.1:7171",
		XpraWSAddress:      "ws://127.0.0.1:2000/",
		Debug:              false,
		ShellCommand:       "bash",
		AddSfUIArgs:        false,
		SfUIOrigin:         "http://127.0.0.1:7171",
		DisableOriginCheck: true,
	}
	sfui.compileClientConfig()

	log.Printf("SFUI [Version : %s] [Built on : %s]\n", "0.1", buildTime)
	log.Printf("Listening on http://%s ....\n", sfui.ServerBindAddress)
	http.ListenAndServe(sfui.ServerBindAddress, http.HandlerFunc(sfui.requestHandler))
}

// Add any UI related configuration that has to be sent to client
// Store it byte format, to prevent json marshalling on every request
// See handleUIConfig()
func (sfui *SfUI) compileClientConfig() {
	compConfig := fmt.Sprintf(
		`{"max_terminals":"%d","auto_login":%s}`,
		sfui.MaxWsTerminals,
		strconv.FormatBool(!sfui.AddSfUIArgs), // Redirect client directly to dashboard if not in global mode.
	)
	sfui.CompiledClientConfig = []byte(compConfig)
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
		sfui.handleWs(w, r)
	case "/xpraws":
		sfui.handleXpraWS(w, r)
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

func (sfui *SfUI) handleWs(w http.ResponseWriter, r *http.Request) {
	clientSecret := r.URL.Query().Get("secret")
	rows, rerr := strconv.ParseUint(r.URL.Query().Get("rows"), 10, 16)
	cols, cerr := strconv.ParseUint(r.URL.Query().Get("cols"), 10, 16)
	if rerr != nil || cerr != nil {
		rows = 30
		cols = 100
	}

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		if !sfui.originAcceptable(ws.Request()) {
			ws.Write([]byte(`unacceptable origin`))
			return
		}

		if err := sfui.secretValid(&TermRequest{
			Secret:   clientSecret,
			ClientIp: r.RemoteAddr,
		}); err != nil { // Invalid Secret
			ws.Write([]byte(err.Error()))
			return
		}

		err := sfui.handleWsPty(&Terminal{
			ClientSecret: clientSecret,
			ClientIp:     r.RemoteAddr,
			WSConn:       ws,
			TermConfig: &TermConfig{
				Rows: uint16(rows),
				Cols: uint16(cols),
			},
		})
		if err != nil {
			ws.Write([]byte(err.Error()))
		}

	}).ServeHTTP(w, r)
}

func (terminal *Terminal) setTermDimensions(rows uint16, cols uint16) {
	if terminal.Pty.Fd() <= 2 {
		return
	}

	window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		uint16(rows),
		uint16(cols),
		0,
		0,
	}
	syscall.Syscall(
		syscall.SYS_IOCTL,
		terminal.Pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
}

// First byte in the chunk sent by the client is a indicator
// of the type of data. This is a custom read implementation
// to handle the first byte.
func (terminal *Terminal) Read(msg []byte) (n int, err error) {
	nmsg := make([]byte, 128)
	n, err = terminal.WSConn.Read(nmsg)
	if n > 0 {
		switch nmsg[0] { // Check the type of data we recieved
		case SFUI_CMD_RESIZE:
			var termConfig TermConfig
			if jerr := json.Unmarshal(nmsg[1:n], &termConfig); jerr == nil {
				terminal.setTermDimensions(termConfig.Rows, termConfig.Cols)
			}
			return 0, nil
		}
		copy(msg, nmsg[1:]) // Copy everything except the first byte
		return n, err
	}
	return n, err
}

var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

func (sfui *SfUI) handleWsPty(terminal *Terminal) error {
	cmdParts := strings.Split(sfui.ShellCommand, " ")
	if sfui.AddSfUIArgs {
		if !isStringAlphabetic(terminal.ClientSecret) {
			return errors.New("unacceptable secret")
		}
		cmdParts = append(cmdParts, fmt.Sprintf(" SECRET=%s", terminal.ClientSecret))
		cmdParts = append(cmdParts, fmt.Sprintf(" REMOTE_ADDR=%s", terminal.ClientIp)) // ClientIP provided by server, no sanitization required
	}

	var err error
	terminal.Pty, err = pty.Start(exec.Command(cmdParts[0], cmdParts[1:]...))
	if err != nil {
		return err
	}
	defer terminal.Pty.Close()

	terminal.setTermDimensions(uint16(terminal.TermConfig.Rows), uint16(terminal.TermConfig.Cols))

	go io.Copy(terminal.WSConn, terminal.Pty) // Copy from PTY -> WS

	// Copy from WS -> PTY, but use the Read() function
	// we defined for Terminal to read from the websocket
	_, werr := io.Copy(terminal.Pty, terminal)
	if werr != nil {
		terminal.WSConn.Close()
	}

	return nil
}

func (sfui *SfUI) secretValid(TermRequest *TermRequest) error {
	// Pass Secret and Client IP to sf Core
	// return errors.New("Banned User")
	// return errors.New("Banned IP")
	return nil
}

func (sfui *SfUI) generateSecret(TermRequest *TermRequest) (Secret string, Error error) {
	// Return a new secret
	return RandomStr(10), nil
}

// Provide UI related config to client
func (sfui *SfUI) handleUIConfig(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(sfui.CompiledClientConfig)
}

func (sfui *SfUI) handleXpraWS(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(sfui.XpraWSAddress)
	wp := websocketproxy.NewProxy(u) // Get rid of this dependency
	wp.Upgrader = websocketproxy.DefaultUpgrader
	wp.Upgrader.CheckOrigin = sfui.originAcceptable
	wp.ServeHTTP(w, r)
}

func (sfui *SfUI) originAcceptable(r *http.Request) bool {
	if !sfui.DisableOriginCheck {
		origin := r.Header.Get("Origin")
		return origin == sfui.SfUIOrigin
	}
	return true
}
