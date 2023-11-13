package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

type TermRequest struct {
	Secret      string `json:"secret"`
	NewInstance bool   `json:"new_instance"`
	ClientIp    string
	TabId       string `json:"tab_id"`
}

type TermResponse struct {
	Status      string `json:"status"`
	Secret      string `json:"secret,omitempty"`
	IsDuplicate bool   `json:"is_duplicate_session,omitempty"`
}

// First byte read from Terminal.Pty is matched with
// the following constants, to determine the type of data
// a client is sending
const (
	SFUI_NORMAL_MSG        = '0'
	SFUI_CMD_RESIZE        = '1'
	SFUI_CMD_PAUSE         = '2'
	SFUI_CMD_RESUME        = '3'
	SFUI_CMD_AUTHENTICATE  = '4'
	SFUI_CMD_PING          = '5'
	SFUI_CMD_PONG          = '6'
	TERM_MAX_AUTH_FAILURES = 3
)

type Terminal struct {
	ClientSecret string
	ClientIp     string
	WSConn       *websocket.Conn
	SSHSession   *ssh.Session
	MsgBuf       []byte
}

type TermConfig struct {
	Secret string `json:"secret"`
	Rows   int    `json:"rows"`
	Cols   int    `json:"cols"`
}

func (sfui *SfUI) handleTerminalWs(w http.ResponseWriter, r *http.Request) {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		clientIp := sfui.getClientAddr(r)

		terminal := Terminal{
			ClientIp: clientIp,
			WSConn:   ws,
			MsgBuf:   make([]byte, 256),
		}

		ws.PayloadType = websocket.BinaryFrame

		clientSecret, cerr := terminal.ReadSecret()
		if cerr != nil {
			ws.Write([]byte(cerr.Error()))
			return
		}
		terminal.ClientSecret = clientSecret

		if !sfui.originAcceptable(ws.Request()) {
			ws.Write([]byte(string(SFUI_NORMAL_MSG) + `unacceptable origin`))
			return
		}

		if err := sfui.secretValid(&TermRequest{
			Secret:   clientSecret,
			ClientIp: clientIp,
		}); err != nil { // Invalid Secret
			ws.Write([]byte(string(SFUI_NORMAL_MSG) + err.Error()))
			return
		}

		err := sfui.handleWsPty(&terminal)
		if err != nil {
			ws.Write([]byte(string(SFUI_NORMAL_MSG) + err.Error()))
		}

	}).ServeHTTP(w, r)
}

func (terminal *Terminal) setTermDimensions(rows int, cols int) {
	terminal.SSHSession.WindowChange(rows, cols)
}

// Read Secret Sent by Client
func (terminal *Terminal) ReadSecret() (secret string, err error) {
	return readSecretFromWs(terminal.WSConn, &terminal.MsgBuf, TERM_MAX_AUTH_FAILURES)
}

func readSecretFromWs(wsConn *websocket.Conn, msgBuf *[]byte, maxAuthFailures int) (secret string, err error) {
	authFailures := 0

	for authFailures < maxAuthFailures {
		n, err := wsConn.Read(*msgBuf)
		if n > 0 && err == nil {
			if (*msgBuf)[0] == SFUI_CMD_AUTHENTICATE { // Check the type of data we recieved
				var termConfig TermConfig
				if jerr := json.Unmarshal((*msgBuf)[1:n], &termConfig); jerr == nil {
					if termConfig.Secret != "" {
						return termConfig.Secret, nil
					}
				}
			}
		}
		authFailures += 1
	}
	return "", fmt.Errorf("Client did not supply valid secret (after %d attempts)", TERM_MAX_AUTH_FAILURES)
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
		case SFUI_CMD_PING:
			terminal.sendPong()
			return 0, nil
		}
		copy(msg, terminal.MsgBuf[1:]) // Copy everything except the first byte
		return n - 1, err
	}
	return n, err
}

var PONG_CMD_BYTES = []byte{SFUI_CMD_PONG} // Mark as Pong
var REG_CMD_BYTES = []byte{'0'}            // Mark as Regular data chunk

func (terminal *Terminal) Write(msg []byte) (n int, err error) {
	bw := append(REG_CMD_BYTES, msg[:]...)
	n, err = terminal.WSConn.Write(bw)
	bw = nil
	return n - 1, err // n-1 so that writer does not get confused as to where the extra 1 byte came from
}

func (terminal *Terminal) sendPong() (n int, err error) {
	return terminal.WSConn.Write(PONG_CMD_BYTES)
}

func (sfui *SfUI) handleWsPty(terminal *Terminal) error {
	if !sfui.ValidSecret(terminal.ClientSecret) {
		return errors.New("unacceptable secret")
	}

	defer sfui.RemoveClientIfInactive(terminal.ClientSecret)

	// Get the  associated client or create a new one
	// client variable below will get stale
	client, cerr := sfui.GetExistingClientOrMakeNew(terminal.ClientSecret, terminal.ClientIp)
	if cerr != nil {
		return cerr
	}

	terr := client.IncTermCount() // Add to terminal  Quota (SFUI.MaxWsTerminals)
	if terr != nil {
		return cerr
	}
	defer client.DecTermCount() // Remove from terminal Quota

	sess, stdin, stdout, stderr, serr := client.SSHConnection.StartTerminal()
	if serr != nil {
		return serr
	}
	terminal.SSHSession = sess
	defer sess.Close()

	stdOutbuf := make([]byte, 32*1024)
	stdErrbuf := make([]byte, 32*1024)

	go io.CopyBuffer(terminal, *stdout, stdOutbuf) // Copy from stdout -> WS
	go io.CopyBuffer(terminal, *stderr, stdErrbuf) // Copy from stderr -> WS

	// Copy from WS -> stdin, but use the Read() function
	// we defined for Terminal to read from the websocket
	done := make(chan error)
	go copyCh(*stdin, terminal, done)

	timeout := time.NewTimer(time.Minute * time.Duration(sfui.WSTimeout))

	select {
	case <-timeout.C:
		break
	case <-done:
		timeout.Stop()
		break
	}

	return nil
}

func (sfui *SfUI) secretValid(TermRequest *TermRequest) error {
	// Pass Secret and Client IP to sf Core
	// return errors.New("Banned User")
	// return errors.New("Banned IP")
	return nil
}

func (sfui *SfUI) generateSecret(TermRequest *TermRequest) string {
	// Return a new secret
	return RandomStr(25)
}

func (sfui *SfUI) originAcceptable(r *http.Request) bool {
	if !sfui.DisableOriginCheck {
		origin := r.Header.Get("Origin")
		return origin == sfui.SfUIOrigin
	}
	return true
}
