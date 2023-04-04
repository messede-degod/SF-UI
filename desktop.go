package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/koding/websocketproxy"
)

func (sfui *SfUI) handleDesktopWS(w http.ResponseWriter, r *http.Request) {
	//Get Secret
	queryVals := r.URL.Query()
	clientSecret := queryVals.Get("secret")

	if clientSecret == "" {
		w.Write([]byte("Invalid Secret"))
		return
	}

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

	werr := sfui.waitForMasterSSHSocket(client.ClientId, time.Second*5, 2)
	if werr != nil {
		w.Write([]byte(werr.Error()))
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
