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

// Reference :  https://raw.githubusercontent.com/pgaskin/easy-novnc/master/server.go

// websockify returns an http.Handler which proxies websocket requests to a VNC server
// address.
func vncWebSockify(to string, viewOnly bool, isSharedConnection bool, closeConnection chan interface{}) http.Handler {
	return websocket.Server{
		Handshake: wsProxyHandshake,
		Handler:   wsProxyHandler(to, viewOnly, isSharedConnection, closeConnection),
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

type ViewOnlyConn struct {
	Ws     *websocket.Conn
	MsgBuf []byte
}

const (
	RFB_KEY_EVENT     = 4
	RFB_POINTER_EVENT = 5
	RFB_CUT_TEXT      = 6
)

// First byte in the chunk is a indicates the type of RFB message.
// https://datatracker.ietf.org/doc/html/rfc6143#section-7.5
func (viewOnlyConn *ViewOnlyConn) Read(msg []byte) (n int, err error) {
	n, err = viewOnlyConn.Ws.Read(viewOnlyConn.MsgBuf)
	if n > 0 {
		switch viewOnlyConn.MsgBuf[0] { // Check the type of message we recieved
		case RFB_CUT_TEXT, RFB_KEY_EVENT, RFB_POINTER_EVENT: // ignore input events
			return 0, nil
		}
		copy(msg, viewOnlyConn.MsgBuf[0:n]) // Copy everything
		return n, err
	}
	return n, err
}

// wsProxyHandler is a websocket.Handler which proxies to a unix address with a
// magic byte check.
func wsProxyHandler(to string, viewOnly bool, isSharedConnection bool, closeConnection chan interface{}) websocket.Handler {
	return func(ws *websocket.Conn) {
		conn, err := net.Dial("unix", to)
		if err != nil {
			log.Println(err)
			ws.Close()
			return
		}

		ws.PayloadType = websocket.BinaryFrame

		done := make(chan error)

		if viewOnly {
			viewOnlyConn := ViewOnlyConn{
				Ws:     ws,
				MsgBuf: make([]byte, 256),
			}
			go copyCh(conn, &viewOnlyConn, done) // Use custom Read() function to filter out input
			go copyCh(ws, conn, done)
		} else {
			go copyCh(conn, ws, done)
			go copyCh(ws, conn, done)
		}

		if isSharedConnection { // if shared, close connection when user disabled sharing(i.e closeConnection channel is closed)
			select {
			case err = <-done:
				if err != nil {
					logf(true, "%v\n", err)
				}
				break
			case _, ok := <-closeConnection:
				if !ok {
					break
				}
			}
		} else { // if not a shared connection, exit only when error occurs
			err = <-done
			if err != nil {
				logf(true, "%v\n", err)
			}
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
