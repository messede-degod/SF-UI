package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func (sfui *SfUI) handleDesktopWS(w http.ResponseWriter, r *http.Request) {
	//Get Secret
	queryVals := r.URL.Query()
	clientSecret := queryVals.Get("secret")
	desktopType := queryVals.Get("type")

	if !sfui.ValidSecret(clientSecret) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`unacceptable secret`))
		return
	}

	defer sfui.RemoveClientIfInactive(clientSecret)

	// Get the  associated client or create a new one
	// client variable below will get stale
	client, cerr := sfui.GetExistingClientOrMakeNew(clientSecret, sfui.getClientAddr(r))
	if cerr != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	defer client.DeactivateDesktopSharing() // Remove all shares when master VNC connection exits

	sfui.startDesktopService(client.MasterSSHConnectionPty, desktopType, time.Second*3)

	vncWebSockify(
		sfui.getGUISocketPath(client.ClientId),
		false, // not view only
		false, // not shared
		client.SharedDesktopConn,
		time.Minute*time.Duration(sfui.WSTimeout),
	).ServeHTTP(w, r)
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

func (sfui *SfUI) handleSharedDesktopWS(w http.ResponseWriter, r *http.Request) {
	// Secret in this case will be the client Id and not the actual secret,
	// this is to prevent the leak of secret to third party.
	queryVals := r.URL.Query()
	clientId := queryVals.Get("client_id")
	shareSecret := queryVals.Get("secret")

	if !sfui.ValidSecret(clientId) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`unacceptable secret`))
		return
	}

	// Get the  associated client
	// client variable below will get stale
	client, cerr := sfui.GetClientById(clientId)
	if cerr != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte(`{"status":"desktop is not active"}`))
		return
	}

	if !client.ShareDesktop && (client.SharedDesktopSecret == shareSecret) {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(`unacceptable secret`))
		return
	}

	serr := client.IncSharedDesktopConnCount()
	if serr != nil {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"status":"maximum shares active"}`))
		return
	}
	defer client.DecSharedDesktopConnCount()

	vncWebSockify(
		sfui.getGUISocketPath(client.ClientId),
		client.SharedDesktopIsViewOnly,
		true, // is a shared connection
		client.SharedDesktopConn,
		time.Minute*time.Duration(sfui.WSTimeout),
	).ServeHTTP(w, r)
}

type DesktopShareRequest struct {
	Secret   string `json:"secret"`
	ClientId string `json:"client_id"`
	Action   string `json:"action"`
	ViewOnly bool   `json:"view_only"`
}

func (sfui *SfUI) handleSetupDesktopSharing(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		desktopShareReq := DesktopShareRequest{}
		if json.Unmarshal(data, &desktopShareReq) == nil {

			var client Client
			var cerr error

			if desktopShareReq.Action == "verify" {
				client, cerr = sfui.GetClientById(desktopShareReq.ClientId)
			} else {
				client, cerr = sfui.GetClient(desktopShareReq.Secret)
			}

			if cerr != nil {
				w.WriteHeader(http.StatusGone)
				w.Write([]byte(`{"status":"desktop is not active"}`))
				return
			}

			if !client.DesktopActive {
				w.WriteHeader(http.StatusGone)
				w.Write([]byte(`{"status":"desktop is not active"}`))
				return
			}

			if client.SharedDesktopConnCount >= client.MaxSharedDesktopConn {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"status":"maximum shares active"}`))
				return
			}

			switch desktopShareReq.Action {
			case "activate":
				sharedSecret := RandomStr(24)
				alreadyShared := client.ActivateDesktopSharing(desktopShareReq.ViewOnly, sharedSecret)
				if alreadyShared {
					sharedSecret = client.SharedDesktopSecret
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{"status":"OK","client_id":"%s","share_secret":"%s"}`,
					client.ClientId, sharedSecret)))
				return
			case "deactivate":
				client.DeactivateDesktopSharing()
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"OK"}`))
				return
			case "verify":
				if client.ShareDesktop && (client.SharedDesktopSecret == desktopShareReq.Secret) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status":"OK"}`))
				} else {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"status":"Desktop Not Shared"}`))
				}
				return
			}
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"Internal Server Error"}`))
}
