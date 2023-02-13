package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

type SfUI struct {
	MaxWsTerminals    int    // Max terminals that can be allocated per client
	ServerBindAddress string // Address to which the current app binds
	DefaultSfEndpoint string // Default endpoint to use for the SSH connections
	ForceSfEndpoint   bool   // Ignore client supplied endpoints
	Debug             bool   // Print debug information
}

type TermRequest struct {
	Secret     string `json:"secret"`
	SfEndpoint string `json:"sf_endpoint"`
}

type TermResponse struct {
	Status     string `json:"status"`
	SfEndpoint string `json:"sf_endpoint"`
}

func main() {
	sfui := SfUI{
		MaxWsTerminals:    5,
		ServerBindAddress: "127.0.0.1:7171",
		DefaultSfEndpoint: "segfault.net",
		ForceSfEndpoint:   true,
		Debug:             true,
	}
	log.Printf("Listening on %s ....\n", sfui.ServerBindAddress)
	http.ListenAndServe(sfui.ServerBindAddress, http.HandlerFunc(sfui.requestHandler))
	return
}

func (sfui *SfUI) requestHandler(w http.ResponseWriter, r *http.Request) {
	if sfui.Debug {
		log.Println(r.RemoteAddr, " ", r.URL, " ", r.UserAgent())
	}
	switch r.URL.Path {
	case "/secret":
		sfui.handleHttp(w, r)
	case "/ws":
		sfui.handleWs(w, r)
	}
}

func (sfui *SfUI) handleHttp(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		termReq := TermRequest{}
		if json.Unmarshal(data, &termReq) == nil {
			if sfui.isSecretValid(&termReq) {
				w.WriteHeader(http.StatusOK)
				termRes := TermResponse{
					Status:     "OK",
					SfEndpoint: sfui.getSfEndpoint(&termReq),
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
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			// Read
			msg := ""
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				return
			}
			// Echo back
			err = websocket.Message.Send(ws, msg)
			if err != nil {
				return
			}
		}
	}).ServeHTTP(w, r)
}

func (sfui *SfUI) isSecretValid(TermRequest *TermRequest) bool {
	return true
}

func (sfui *SfUI) getSfEndpoint(TermRequest *TermRequest) string {
	if sfui.ForceSfEndpoint || TermRequest.SfEndpoint == "" ||
		!validSfEndpoint(TermRequest.SfEndpoint) {
		return sfui.DefaultSfEndpoint
	}
	return TermRequest.SfEndpoint
}

func validSfEndpoint(Endpoint string) bool {
	return true
}
