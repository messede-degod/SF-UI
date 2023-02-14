package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"golang.org/x/net/websocket"
)

type SfUI struct {
	MaxWsTerminals    int    // Max terminals that can be allocated per client
	ServerBindAddress string // Address to which the current app binds
	Debug             bool   // Print debug information
	ShellCommand      string // Command to run on the pty
	// Two additional arguments are added to ShellCommand
	// 	Example:  somecmd  SECRET=abc REMOTE_IP=1.1.1.1
	AddSfUIArgs          bool
	CompiledClientConfig []byte // Ui related onfig that has to be sent to client
	// SfEndpoint           string  // Current Sf Endpoints Name

}

type TermRequest struct {
	Secret   string `json:"secret"`
	ClientIp string
	// SfEndpoint string `json:"sf_endpoint"`
}

type TermResponse struct {
	Status string `json:"status"`
	// SfEndpoint string `json:"sf_endpoint"`
}

type Terminal struct {
	ClientSecret string
	ClientIp     string
	WSConn       *websocket.Conn
	Rows         uint16
	Cols         uint16
}

var buildTime string

//go:embed ui/dist/sf-ui
var staticfiles embed.FS

func main() {
	sfui := SfUI{
		MaxWsTerminals:    10,
		ServerBindAddress: "127.0.0.1:7171",
		Debug:             false,
		ShellCommand:      "sshpass -p segfault ssh root@segfault.net",
		AddSfUIArgs:       false,
	}
	sfui.compileClientConfig()

	log.Printf("SFUI [Version : %s] [Built on : %s]\n", "0.1", buildTime)
	log.Printf("Listening on http://%s ....\n", sfui.ServerBindAddress)
	http.ListenAndServe(sfui.ServerBindAddress, http.HandlerFunc(sfui.requestHandler))
	return
}

// Add any UI related configuration that has to be sent to client
// Store it byte format, to prevent json marshalling on every request
// See handleUIConfig()
func (sfui *SfUI) compileClientConfig() {
	compConfig := fmt.Sprintf(`{"max_terminals":"%d"}`, sfui.MaxWsTerminals)
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
	case "/ws":
		sfui.handleWs(w, r)
		return // Dont add json header to WS requests
	case "/config":
		sfui.handleUIConfig(w, r)
	default:
		handleUIRequest(w, r)
		return // Dont add json header to UI requests
	}
	w.Header().Add("Content-Type", "application/json")
}

func (sfui *SfUI) handleSecret(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		termReq := TermRequest{}
		if json.Unmarshal(data, &termReq) == nil {
			termReq.ClientIp = r.RemoteAddr
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
			Rows:         uint16(rows),
			Cols:         uint16(cols),
		})
		if err != nil {
			ws.Write([]byte(err.Error()))
		}

	}).ServeHTTP(w, r)
}

func setTermDimensions(rows uint16, cols uint16, fd uintptr) {
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
		fd,
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
}

var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

func (sfui *SfUI) handleWsPty(terminal *Terminal) error {
	cmdParts := strings.Split(sfui.ShellCommand, " ")
	if !isStringAlphabetic(terminal.ClientSecret) {
		return errors.New("Unacceptable Secret")
	}
	cmdParts = append(cmdParts, fmt.Sprintf(" SECRET=%s", terminal.ClientSecret))
	cmdParts = append(cmdParts, fmt.Sprintf(" REMOTE_IP=%s", terminal.ClientIp)) // ClientIP provided by server, no sanitization required

	apty, err := pty.Start(exec.Command(cmdParts[0], cmdParts[1:]...))
	if err != nil {
		return err
	}
	defer apty.Close()

	setTermDimensions(uint16(terminal.Rows), uint16(terminal.Cols), apty.Fd())

	go func() {
		for {
			_, rerr := io.Copy(terminal.WSConn, apty)
			if rerr != nil {
				break
			}
		}
	}()

	_, werr := io.Copy(apty, terminal.WSConn)
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

// Provide UI related config to client
func (sfui *SfUI) handleUIConfig(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(sfui.CompiledClientConfig)
}
