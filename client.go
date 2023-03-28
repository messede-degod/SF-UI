package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os/exec"
	"sync"
	"time"
)

type Client struct {
	ClientId               string
	mu                     *sync.Mutex
	TerminalsCount         int
	MasterSSHConnectionCmd *exec.Cmd
	DesktopActive          bool
	MaxTerms               int
}

var clients = make(map[string]Client)
var randVal = RandomStr(10) // Random str for derving clientId, doesnt change unless sfui is restarted

// Return a new client, prepare necessary sockets
func (sfui *SfUI) NewClient(ClientSecret string, ClientIp string) (Client, error) {
	// Make and return a new client
	client := Client{
		ClientId:       getClientId(ClientSecret),
		mu:             &sync.Mutex{},
		TerminalsCount: 0,
		MaxTerms:       sfui.MaxWsTerminals,
	}

	if werr := sfui.workDirAddClient(client.ClientId); werr != nil {
		return client, werr
	}

	mcCmd, merr := sfui.prepareMasterSSHSocket(client.ClientId, ClientSecret, ClientIp)
	if merr != nil {
		return client, merr
	}
	client.MasterSSHConnectionCmd = mcCmd

	// Wait untill the master SSH socket becomes available
	mwerr := sfui.waitForMasterSSHSocket(client.ClientId, time.Second*10, 2)
	if mwerr != nil {
		return client, mwerr
	}

	if gerr := sfui.prepareWsBridgeSocket(client.ClientId, ClientSecret, ClientIp); gerr != nil {
		return client, gerr
	}

	clients[client.ClientId] = client

	return client, nil
}

func (sfui *SfUI) RemoveClient(client *Client) {
	sfui.destroyMasterSSHSocket(client)
	sfui.workDirRemoveClient(client.ClientId)
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
	client, ok := clients[getClientId(ClientSecret)]
	if ok {
		return client, nil
	}
	return sfui.NewClient(ClientSecret, ClientIp)
}

func (sfui *SfUI) GetClient(ClientSecret string) (Client, error) {
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
