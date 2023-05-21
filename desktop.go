package main

import (
	"net/http"
	"os"
	"time"
)

func (sfui *SfUI) handleDesktopWS(w http.ResponseWriter, r *http.Request) {
	//Get Secret
	queryVals := r.URL.Query()
	clientSecret := queryVals.Get("secret")
	desktopType := queryVals.Get("type")

	if !validSecret(clientSecret) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`unacceptable secret`))
		return
	}

	defer sfui.RemoveClientIfInactive(clientSecret)

	// Get the  associated client or create a new one
	// client variable below will get stale
	client, cerr := sfui.GetExistingClientOrMakeNew(clientSecret, sfui.getClientAddr(r))
	if cerr != nil {
		w.Write([]byte(cerr.Error()))
		return
	}

	if client.DesktopIsActivate() {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(`can only have one desktop connection active at a time`))
		return
	}

	client.ActivateDesktop()
	defer client.DeActivateDesktop()

	sfui.startDesktopService(client.MasterSSHConnectionPty, desktopType, time.Second*3)

	vncWebSockify(sfui.getGUISocketPath(client.ClientId), false).ServeHTTP(w, r)
}

// Issue appropriate desktop start command(Type) using Pty and Wait for a certain duration
// so that the said command can come alive
func (sfui *SfUI) startDesktopService(Pty *os.File, Type string, Wait time.Duration) {
	startCmd := ""
	switch Type {
	case "xpra":
		startCmd = sfui.StartXpraCommand
	default:
		startCmd = sfui.StartVNCCommand
	}
	Pty.WriteString(startCmd)
	Pty.Sync()
	time.Sleep(Wait)
}
