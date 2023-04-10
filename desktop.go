package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/koding/websocketproxy"
)

func (sfui *SfUI) handleDesktopWS(w http.ResponseWriter, r *http.Request) {
	//Get Secret
	queryVals := r.URL.Query()
	clientSecret := queryVals.Get("secret")

	if !validSecret(clientSecret) {
		w.Write([]byte(`unacceptable secret`))
		return
	}

	// Get the  associated client or create a new one
	// client variable below will get stale
	client, cerr := sfui.GetExistingClientOrMakeNew(clientSecret, sfui.getClientAddr(r))
	if cerr != nil {
		w.Write([]byte(cerr.Error()))
		return
	}

	if client.DesktopIsActivate() {
		w.Write([]byte(`can only have one desktop connection active at a time`))
		return
	}

	client.ActivateDesktop()
	defer sfui.RemoveClientIfInactive(clientSecret)
	defer client.DeActivateDesktop()

	u, _ := url.Parse("unix://" + sfui.getGUISocketPath(client.ClientId))
	wp := websocketproxy.NewUnixProxy(u) // Get rid of this dependency
	wp.Upgrader = websocketproxy.DefaultUpgrader
	wp.Upgrader.CheckOrigin = sfui.originAcceptable
	wp.ServeHTTP(w, r)
}

type setupDesktop struct {
	DesktopType  string `json:"desktop_type"` // xpra,novnc
	ClientSecret string `json:"client_secret"`
}

// start the GUI service on the instance(ex: startxweb), use the master connection
// to issue commands.
func (sfui *SfUI) handleSetupDesktop(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		setupDesktopReq := setupDesktop{}
		if json.Unmarshal(data, &setupDesktopReq) == nil {
			if !validSecret(setupDesktopReq.ClientSecret) {
				w.Write([]byte(`unacceptable secret`))
				return
			}

			// Get the  associated client or create a new one
			// client variable below will get stale
			client, cerr := sfui.GetExistingClientOrMakeNew(setupDesktopReq.ClientSecret,
				strings.Split(r.RemoteAddr, ":")[0])
			if cerr != nil {
				w.Write([]byte(cerr.Error()))
				return
			}

			if client.DesktopIsActivate() {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"status":"desktop connection is already active for client"}`))
				return
			}

			startCmd := ""
			switch setupDesktopReq.DesktopType {
			case "novnc":
				startCmd = sfui.StartNoVNCCommand
			default:
				startCmd = sfui.StartXpraCommand
			}

			// Check for short writes
			client.MasterSSHConnectionPty.WriteString(startCmd)
			client.MasterSSHConnectionPty.Sync()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"OK"}`))
			return
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"Internal Server Error"}`))
}
