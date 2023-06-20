package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var proxy, perr = NewHttpToUnixProxy("/tmp/test.sock")

func (sfui *SfUI) handleFileBrowser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}

	clientSecret := r.Header.Get("X-SfUi-Token")
	if clientSecret == "" {
		clientSecret = r.URL.Query().Get("sf-secret")
	}

	if !sfui.ValidSecret(clientSecret) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"status":"Invalid Secret"}`))
		return
	}

	client, err := sfui.GetClient(clientSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"status":"%s"}`, err.Error())))
		return
	}

	if client.FileBrowserProxy == nil { // Proxy can timeout after certain time
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte(fmt.Sprintf(`{"status":"%s"}`, err.Error())))
		return
	}

	r.URL.Path = strings.Replace(r.URL.Path, "/filebrowser", "", 1)
	client.FileBrowserProxy.ServeHTTP(w, r)
}

type setupFileBrowser struct {
	DesktopType  string `json:"desktop_type"` // xpra,novnc
	ClientSecret string `json:"client_secret"`
}

// start the GUI service on the instance(ex: startfb), use the master connection
// to issue commands.
func (sfui *SfUI) handleSetupFileBrowser(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		setupFileBrowserReq := setupFileBrowser{}
		if json.Unmarshal(data, &setupFileBrowserReq) == nil {
			if !sfui.ValidSecret(setupFileBrowserReq.ClientSecret) {
				w.Write([]byte(`unacceptable secret`))
				return
			}

			// Get the  associated client or create a new one
			// client variable below will get stale
			client, cerr := sfui.GetClient(setupFileBrowserReq.ClientSecret)
			if cerr != nil {
				w.WriteHeader(http.StatusUnavailableForLegalReasons)
				w.Write([]byte(fmt.Sprintf(`{"status":"%s"}`, cerr.Error())))
				return
			}

			if !client.MasterSSHConnectionActive {
				werr := sfui.waitForMasterSSHSocket(client.ClientId, 5, 2)
				if werr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf(`{"status":"%s"}`, werr.Error())))
					return
				}

				// serialize access to client, since it can be updated inbetween by NewClient()
				client.mu.Lock()
				defer client.mu.Unlock()

				// master SSH socket is now active, grab a fresh copy of the client
				client, cerr = sfui.GetClient(setupFileBrowserReq.ClientSecret)
			}

			// TODO : Check for short writes
			client.MasterSSHConnectionPty.WriteString(sfui.StartFileBrowserCommand)
			client.MasterSSHConnectionPty.Sync()

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"OK"}`))
			return
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"Internal Server Error"}`))
}
