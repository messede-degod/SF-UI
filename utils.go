package main

import (
	"crypto/rand"
	"math/big"
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
