package main

import (
	"crypto/rand"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func RandomStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n-1)
	for i := range s {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err == nil {
			s[i] = letters[n.Int64()]
		}
	}

	return string(s)
}

func (sfui *SfUI) getClientAddr(r *http.Request) string {
	if sfui.UseXForwardedForHeader {
		fwAddr := r.Header.Get("X-Forwarded-For")
		if fwAddr != "" {
			return strings.Replace(fwAddr, ":", "", -1) // Go returns address in v6 compatibility mode ':'
		}
	}
	return strings.Split(r.RemoteAddr, ":")[0] // Remote addr is ip:port, we need only ip
}

// NewHttpToUnixProxy : Proxy between http/websocket and unix socket
func NewHttpToUnixProxy(sockAddr string) (*httputil.ReverseProxy, error) {
	// Traget Url value is never used since we use the unix transport, but it has to speicifed anyhow
	targetUrl := "http://127.0.0.1:8080"
	target, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &http.Transport{
		Dial: func(proto, addr string) (conn net.Conn, err error) {
			return net.Dial("unix", sockAddr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return proxy, nil
}
