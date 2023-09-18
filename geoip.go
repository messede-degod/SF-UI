package main

import (
	"errors"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

var db *maxminddb.Reader
var IsActive bool

func GeoIpInit(FilePath string) error {
	var err error
	db, err = maxminddb.Open(FilePath)
	if err == nil {
		IsActive = true
	}
	return err
}

func GeoIpClose() error {
	return db.Close()
}

func GeoIpLookup(IP string) (string, error) {
	if IsActive {
		ip := net.ParseIP(IP)

		var record struct {
			Country struct {
				ISOCode string `maxminddb:"iso_code"`
			} `maxminddb:"country"`
		}

		err := db.Lookup(ip, &record)

		return record.Country.ISOCode, err
	}
	return "", errors.New("MMDB Uninitialized")
}
