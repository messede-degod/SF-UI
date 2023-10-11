package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type SFBanDB struct {
	Addresses map[string]string `json:"addresses"` // IP->reason
}

var BanDB = SFBanDB{}

func (banDB *SFBanDB) Init() {
	banDB.Addresses = make(map[string]string)
	data, err := os.ReadFile("banDB.json")
	if err == nil {
		json.Unmarshal(data, banDB)
	}
}

func (banDB *SFBanDB) Save() {
	banData, err := json.Marshal(BanDB)
	if err == nil {
		os.WriteFile("banDB.json", banData, 0644)
	}
}

func (banDB *SFBanDB) IsBanned(ip string) (isBanned bool, reason string) {
	if banDB.Addresses != nil {
		reason, isBanned = banDB.Addresses[ip]
		return isBanned, reason
	}
	return false, ""
}

type BanDBOp struct {
	Ip     string `json:"ip"`
	Reason string `json:"reason"`
}

func (sfui *SfUI) AddBan(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	MtSecret := r.Header.Get("X-Mt-Secret")

	if MtSecret != sfui.MaintenanceSecret {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"denied"}`))
		return
	}

	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		banRequest := BanDBOp{}
		if json.Unmarshal(data, &banRequest) == nil {
			if banRequest.Ip != "" && banRequest.Reason != "" && BanDB.Addresses != nil {
				BanDB.Addresses[banRequest.Ip] = banRequest.Reason
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
				return
			}
		}
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"error"}`))
}

func (sfui *SfUI) RemoveBan(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	MtSecret := r.Header.Get("X-Mt-Secret")

	if MtSecret != sfui.MaintenanceSecret {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"denied"}`))
		return
	}

	data, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err == nil {
		banRequest := BanDBOp{}
		if json.Unmarshal(data, &banRequest) == nil {
			if banRequest.Ip != "" && BanDB.Addresses != nil {
				delete(BanDB.Addresses, banRequest.Ip)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
				return
			}
		}
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"error"}`))
}

func (sfui *SfUI) ListBans(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	MtSecret := r.Header.Get("X-Mt-Secret")

	if MtSecret != sfui.MaintenanceSecret {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"denied"}`))
		return
	}

	banData, err := json.Marshal(BanDB)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write(banData)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"status":"error"}`))
}
