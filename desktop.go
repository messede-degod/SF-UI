package main

import (
	"net/http"
	"net/url"

	"github.com/koding/websocketproxy"
)

func (sfui *SfUI) handleDesktopWS(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(sfui.XpraWSAddress)
	wp := websocketproxy.NewProxy(u) // Get rid of this dependency
	wp.Upgrader = websocketproxy.DefaultUpgrader
	wp.Upgrader.CheckOrigin = sfui.originAcceptable
	wp.ServeHTTP(w, r)
}
