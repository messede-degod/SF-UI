package main

import (
	"mime"
	"net/http"
	"strings"
)

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
