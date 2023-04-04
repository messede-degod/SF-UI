package main

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strings"
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
	if sfui.UseXForwaredForHeader {
		fwAddr := r.Header.Get("X-Forwared-For")
		if fwAddr != "" {
			return fwAddr
		}
	}
	return strings.Split(r.RemoteAddr, ":")[0] // Remote addr is ip:port, we need only ip
}
