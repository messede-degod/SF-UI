package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

// Borrowed from https://raw.githubusercontent.com/pgaskin/easy-novnc/master/server.go

// websockify returns an http.Handler which proxies websocket requests to a tcp
// address and checks magic bytes.
func websockify(to string, magic []byte) http.Handler {
	return websocket.Server{
		Handshake: wsProxyHandshake,
		Handler:   wsProxyHandler(to, magic),
	}
}

// wsProxyHandshake is a handshake handler for a websocket.Server.
func wsProxyHandshake(config *websocket.Config, r *http.Request) error {
	if r.Header.Get("Sec-WebSocket-Protocol") != "" {
		config.Protocol = []string{"binary"}
	}
	r.Header.Set("Access-Control-Allow-Origin", "*")
	r.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
	return nil
}

// wsProxyHandler is a websocket.Handler which proxies to a unix address with a
// magic byte check.
func wsProxyHandler(to string, magic []byte) websocket.Handler {
	return func(ws *websocket.Conn) {
		conn, err := net.Dial("unix", to)
		if err != nil {
			log.Println(err)
			ws.Close()
			return
		}

		ws.PayloadType = websocket.BinaryFrame

		m := newMagicCheck(conn, magic)

		done := make(chan error)
		go copyCh(conn, ws, done)
		go copyCh(ws, m, done)

		err = <-done
		if m.Failed() {
			logf(true, "attempt to connect to non-VNC port (%s, %#v)\n", to, string(m.Magic()))
		} else if err != nil {
			logf(true, "%v\n", err)
		}

		conn.Close()
		ws.Close()
		<-done
	}
}

func logf(cond bool, format string, a ...interface{}) {
	if cond {
		fmt.Printf("%s: %s", time.Now().Format("Jan 02 15:04:05"), fmt.Sprintf(format, a...))
	}
}

// copyCh is like io.Copy, but it writes to a channel when finished.
func copyCh(dst io.Writer, src io.Reader, done chan error) {
	_, err := io.Copy(dst, src)
	done <- err
}

// magicCheck implements an efficient wrapper around an io.Reader which checks
// for magic bytes at the beginning, and will return a sticky io.EOF and stop
// reading from the original reader as soon as a mismatch starts.
type magicCheck struct {
	rdr io.Reader
	exp []byte
	len int
	rem int
	act []byte
	fld bool
}

func newMagicCheck(r io.Reader, magic []byte) *magicCheck {
	return &magicCheck{r, magic, len(magic), len(magic), make([]byte, len(magic)), false}
}

// Failed returns true if the magic check has failed (note that it returns false
// if the source io.Reader reached io.EOF before the check was complete).
func (m *magicCheck) Failed() bool {
	return m.fld
}

// Magic returns the magic which was read so far.
func (m *magicCheck) Magic() []byte {
	return m.act
}

func (m *magicCheck) Read(buf []byte) (n int, err error) {
	if m.fld {
		return 0, io.EOF
	}
	n, err = m.rdr.Read(buf)
	if err == nil && n > 0 && m.rem > 0 {
		m.rem -= copy(m.act[m.len-m.rem:], buf[:n])
		for i := 0; i < m.len-m.rem; i++ {
			if m.act[i] != m.exp[i] {
				m.fld = true
				return 0, io.EOF
			}
		}
	}
	return n, err
}
