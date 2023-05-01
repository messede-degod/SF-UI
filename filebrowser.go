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

	if validSecret(clientSecret) {
		client, err := sfui.GetExistingClientOrMakeNew(clientSecret,
			strings.Split(r.RemoteAddr, ":")[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"status":"%s"}`, err.Error())))
			return
		}

		r.URL.Path = strings.Replace(r.URL.Path, "/filebrowser", "", 1)
		client.FileBrowserProxy.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"status":"Invalid Secret"}`))
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
			if !validSecret(setupFileBrowserReq.ClientSecret) {
				w.Write([]byte(`unacceptable secret`))
				return
			}

			// Get the  associated client or create a new one
			// client variable below will get stale
			client, cerr := sfui.GetExistingClientOrMakeNew(setupFileBrowserReq.ClientSecret,
				strings.Split(r.RemoteAddr, ":")[0])
			if cerr != nil {
				w.Write([]byte(cerr.Error()))
				return
			}

			if client.FileBrowserServiceIsActivate() {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"status":"filebrowser service is already active for client"}`))
				return
			}

			// TODO : Check for short writes
			client.MasterSSHConnectionPty.WriteString("startfb")
			client.MasterSSHConnectionPty.Sync()

			client.ActivateFileBrowserService()

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"OK"}`))
			return
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"Internal Server Error"}`))
}
