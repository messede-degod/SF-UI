package main

import (
	"net/http"
	"regexp"
)

var routes map[string]func(w http.ResponseWriter, r *http.Request)
var isFbPath = regexp.MustCompile(`(?m)^/filebrowser.*`).MatchString

func (sfui *SfUI) InitRouter() {
	routes = map[string]func(w http.ResponseWriter, r *http.Request){
		"/secret":          sfui.handleLogin, // login
		"/logout":          sfui.handleLogout,
		"/config":          sfui.handleUIConfig,
		"/ws":              sfui.handleTerminalWs,
		"/desktopws":       sfui.handleDesktopWS,
		"/sharedDesktopWs": sfui.handleSharedDesktopWS,
		"/filebrowser":     sfui.handleSetupFileBrowser,
		"/desktop/share":   sfui.handleSetupDesktopSharing,
		"/stats":           sfui.handleClientStats,
		"/ban/add":         sfui.AddBan,
		"/ban/remove":      sfui.RemoveBan,
		"/ban/list":        sfui.ListBans,
	}
}

func (sfui *SfUI) requestHandler(w http.ResponseWriter, r *http.Request) {
	if sfui.Debug {
		// log.Println(r.RemoteAddr, " ", r.URL, " ", r.UserAgent())
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "*")
		w.Header().Add("Access-Control-Allow-Headers", "*")
	}

	if handler, ok := routes[r.URL.Path]; ok {
		handler(w, r)
		return
	}

	// /filebrowser/*
	if isFbPath(r.URL.Path) {
		sfui.handleFileBrowser(w, r)
		return
	}

	handleUIRequest(w, r)
}
