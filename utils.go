package main

import (
	"crypto/rand"
	"io"
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
			fwAddr = strings.ReplaceAll(fwAddr, ":", "") // Go returns address in v6 compatibility mode ':'
			return strings.ReplaceAll(fwAddr, "f", "")   // remove 'f's
		}
	}
	return strings.Split(r.RemoteAddr, ":")[0] // Remote addr is ip:port, we need only ip
}

// NewHttpToNetConnProxy : Proxy between http/websocket and net.Conn
func NewHttpToNetConnProxy(hconn *net.Conn) (*httputil.ReverseProxy, error) {
	// Target Url value is never used since we use a custom Dial, but it has to specified anyhow
	targetUrl := "http://127.0.0.1:8080"
	target, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &http.Transport{
		Dial: func(proto, addr string) (conn net.Conn, err error) {
			return *hconn, nil
		},
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return proxy, nil
}

// copyCh is like io.Copy, but it writes to a channel when finished.
func copyCh(dst io.Writer, src io.Reader, done chan error) {
	buf := make([]byte, 32*1024)
	_, err := io.CopyBuffer(dst, src, buf)
	done <- err
}

func GetCountryByIp(ip string) string {
	countryCode, err := GeoIpLookup(ip)
	if err != nil {
		return "WORLD"
	}
	if countryCode == "" {
		return "LOCAL"
	}
	return countryCode
}
