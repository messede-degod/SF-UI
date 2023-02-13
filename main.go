package main

import (
	"embed"
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

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

var buildTime string

//go:embed ui/dist/sf-ui
var staticfiles embed.FS

func main() {
	sfui := SfUI{
		MaxWsTerminals:    5,
		ServerBindAddress: "127.0.0.1:7171",
		DefaultSfEndpoint: "segfault.net",
		ForceSfEndpoint:   true,
		Debug:             true,
	}
	log.Printf("SFUI [Version : %s] [Built on : %s]\n", "0.1", buildTime)
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
	default:
		handleUIRequest(w, r)
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
				w.Header().Add("Content-Type", "application/json")
				w.Write(response)
				return
			}
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Add("Content-Type", "application/json")
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

func handleUIRequest(w http.ResponseWriter, r *http.Request) {
	pagePrefix := "ui/dist/sf-ui"
	var page string

	// Redirect / to /index.html
	if r.URL.Path == "/" {
		page = pagePrefix + "/index.html"
	} else {
		page = pagePrefix + r.URL.Path
	}

	// Enable Caching for everything other than index.html
	if page != pagePrefix+"/index.html" {
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
		w.Header().Add("Cache-Control", "public max-age=31535996 immutable")
	}

	// Read the requested file from the FS
	fileBytes, err := staticfiles.ReadFile(page)
	if err == nil {
		w.Header().Add("Content-Type", getContentType(&page))
		w.Header().Add("Last-Modified", buildTime)
		w.Write(fileBytes)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("404 Not Found"))
}

// Given a file name return the appropriate content type
func getContentType(filename *string) string {
	splits := strings.Split(*filename, ".")
	return mime.TypeByExtension("." + splits[len(splits)-1])
}
