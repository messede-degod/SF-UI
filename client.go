package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Client struct {
	ClientId                  string
	mu                        *sync.Mutex
	TerminalsCount            int
	MasterSSHConnectionActive bool
	MasterSSHConnectionCmd    *exec.Cmd
	MasterSSHConnectionPty    *os.File
	DesktopActive             bool
	DesktopServiceActive      bool
	MaxTerms                  int
}

var clients = make(map[string]Client) // Clients DB
var cmu = &sync.Mutex{}               // Synchronize access to the clients DB above
var randVal = RandomStr(10)           // Random str for deriving clientId, doesnt change unless sfui is restarted

// Return a new client, prepare necessary sockets
func (sfui *SfUI) NewClient(ClientSecret string, ClientIp string) (Client, error) {
	// Make and return a new client
	client := Client{
		ClientId:                  getClientId(ClientSecret),
		mu:                        &sync.Mutex{},
		TerminalsCount:            0,
		MasterSSHConnectionActive: false,
		MaxTerms:                  sfui.MaxWsTerminals,
	}

	// Make a inital entry in the clients DB, this is to prevent a race condition
	// where multiple SSH connection would be created when a master SSH connection
	// is still being established.
	cmu.Lock()
	clients[client.ClientId] = client
	cmu.Unlock()

	if werr := sfui.workDirAddClient(client.ClientId); werr != nil {
		return client, werr
	}

	mCmd, mPty, mPtyErr := sfui.prepareMasterSSHSocket(client.ClientId, ClientSecret, ClientIp)
	if mPtyErr != nil {
		return client, mPtyErr
	}
	client.MasterSSHConnectionPty = mPty
	client.MasterSSHConnectionCmd = mCmd

	// Wait untill the master SSH socket becomes available
	mwerr := sfui.waitForMasterSSHSocket(client.ClientId, time.Second*10, 2)
	if mwerr != nil {
		return client, mwerr
	}

	cmu.Lock()
	client.MasterSSHConnectionActive = true
	clients[client.ClientId] = client
	cmu.Unlock()

	return client, nil
}

func (sfui *SfUI) RemoveClient(client *Client) {
	cmu.Lock()
	defer cmu.Unlock()

	sfui.destroyMasterSSHSocket(client)
	wrerr := sfui.workDirRemoveClient(client.ClientId)
	if wrerr != nil {
		log.Println(wrerr)
	}
	delete(clients, client.ClientId)
}

// If a client has no active terminals or a GUI connection
// consider them as inactive and tear down the master SSH connection
func (sfui *SfUI) RemoveClientIfInactive(clientSecret string) {
	// Obtain a fresh copy of the client
	client, err := sfui.GetClient(clientSecret)
	if err == nil {
		client.mu.Lock()
		if client.TerminalsCount == 0 && !client.DesktopActive {
			sfui.RemoveClient(&client)
		} else { // If we removed client in the previous block there will be nothing left to unlock
			client.mu.Unlock()
		}
	}
}

func (sfui *SfUI) GetExistingClientOrMakeNew(ClientSecret string, ClientIp string) (Client, error) {
	client, cerr := sfui.GetClient(ClientSecret)
	if cerr != nil {
		return sfui.NewClient(ClientSecret, ClientIp)
	}

	if !client.MasterSSHConnectionActive {
		werr := sfui.waitForMasterSSHSocket(client.ClientId, 5, 2)
		if werr != nil {
			return client, werr
		}
		// master SSH socket is now active, grab a fresh copy of the client
		client, cerr = sfui.GetClient(ClientSecret)
	}

	return client, cerr
}

func (sfui *SfUI) GetClient(ClientSecret string) (Client, error) {
	cmu.Lock()
	defer cmu.Unlock()

	client, ok := clients[getClientId(ClientSecret)]
	if ok {
		return client, nil
	}
	return client, errors.New("No such client")
}

// Derive a client id from secret
func getClientId(ClientSecret string) string {
	h := sha256.New()
	h.Write([]byte(ClientSecret))
	h.Write([]byte(randVal))
	return hex.EncodeToString(h.Sum(nil))
}

func (client *Client) IncTermCount() error {
	// mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]

	if fclient.TerminalsCount >= fclient.MaxTerms {
		return errors.New("Max Terminals allocated")
	}
	fclient.TerminalsCount += 1

	// update client details
	clients[client.ClientId] = fclient
	return nil
}

func (client *Client) DecTermCount() {
	// mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]

	if fclient.TerminalsCount > 0 {
		fclient.TerminalsCount -= 1
	}

	clients[client.ClientId] = fclient
}

func (client *Client) ActivateDesktop() {
	// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]
	fclient.DesktopActive = true
	clients[client.ClientId] = fclient
}

func (client *Client) DeActivateDesktop() {
	// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]
	fclient.DesktopActive = false
	clients[client.ClientId] = fclient
}

func (client *Client) DesktopIsActivate() bool {
	// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]
	return fclient.DesktopActive
}

func (client *Client) ActivateDesktopService() {
	// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]
	fclient.DesktopServiceActive = true
	clients[client.ClientId] = fclient
}

func (client *Client) DesktopServiceIsActivate() bool {
	// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]
	return fclient.DesktopServiceActive
}
