package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"golang.org/x/net/websocket"
)

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
	MsgBuf       []byte
}

type TermConfig struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

func (sfui *SfUI) handleTerminalWs(w http.ResponseWriter, r *http.Request) {
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
			MsgBuf: make([]byte, 256),
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
	n, err = terminal.WSConn.Read(terminal.MsgBuf)
	if n > 0 {
		switch terminal.MsgBuf[0] { // Check the type of data we recieved
		case SFUI_CMD_RESIZE:
			var termConfig TermConfig
			if jerr := json.Unmarshal(terminal.MsgBuf[1:n], &termConfig); jerr == nil {
				terminal.setTermDimensions(termConfig.Rows, termConfig.Cols)
			}
			return 0, nil
		}
		copy(msg, terminal.MsgBuf[1:]) // Copy everything except the first byte
		return n - 1, err
	}
	return n, err
}

var validSecret = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

func (sfui *SfUI) handleWsPty(terminal *Terminal) error {
	shellCommand := sfui.ShellCommand
	if sfui.AddSfUIArgs {
		if !validSecret(terminal.ClientSecret) {
			return errors.New("unacceptable secret")
		}
		if strings.Count(sfui.ShellCommand, "]s") >= 1 { // trying to match %[1]s and %[2]s
			shellCommand = fmt.Sprintf(sfui.ShellCommand, terminal.ClientSecret, terminal.ClientIp)
		}
	}

	var err error
	command := exec.Command("bash", "-c", shellCommand)
	terminal.Pty, err = pty.Start(command)
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

	command.Process.Kill()
	command.Wait()
	terminal.Pty.Close()
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
	return RandomStr(25), nil
}

func (sfui *SfUI) originAcceptable(r *http.Request) bool {
	if !sfui.DisableOriginCheck {
		origin := r.Header.Get("Origin")
		return origin == sfui.SfUIOrigin
	}
	return true
}
