package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http/httputil"
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
	DesktopActive             bool // Whether a active desktop ws connection exists
	MaxTerms                  int
	MaxSharedDesktopConn      int
	FileBrowserProxy          *httputil.ReverseProxy
	FileBrowserServiceActive  bool
	ShareDesktop              bool // Whether desktop sharing is active
	SharedDesktopIsViewOnly   bool
	SharedDesktopSecret       string
	SharedDesktopConn         chan interface{} // Channel when closed kills all shared desktop connections
	SharedDesktopConnCount    int              // No of active connections to shared desktop
	// Channel when closed prevents master SSH connection from being killed by RemoveClientIfInactive,
	// that is unless a ClientInactivityTimeout is first reached, open channel indicated a inactive client
	// closed channel indicates a active client
	ClientConn   chan interface{}
	ClientActive bool // Atleast one active connection exists
	// Random value supplied by client during login , helps to identify duplicate sessions
	TabId string
}

var AcceptClients = true
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
		MaxSharedDesktopConn:      sfui.MaxSharedDesktopConn,
		ClientConn:                make(chan interface{}), // Initially no active connections exist
		ClientActive:              true,
	}

	if !AcceptClients {
		return client, errors.New("Not Accepting New Clients !")
	}

	// Prevent use of a unprepared client
	client.mu.Lock()
	defer client.mu.Unlock()

	// Make a inital entry in the clients DB, this is to prevent a race condition
	// where multiple SSH connection would be created when a master SSH connection
	// is still being established.
	cmu.Lock()
	clients[client.ClientId] = client
	cmu.Unlock()

	if werr := sfui.workDirAddClient(client.ClientId); werr != nil {
		sfui.RemoveClient(&client)
		return client, werr
	}

	mCmd, mPty, mPtyErr := sfui.prepareMasterSSHSocket(client.ClientId, ClientSecret, ClientIp)
	if mPtyErr != nil {
		sfui.RemoveClient(&client)
		return client, mPtyErr
	}
	client.MasterSSHConnectionPty = mPty
	client.MasterSSHConnectionCmd = mCmd

	// Wait untill the master SSH socket becomes available
	mwerr := sfui.waitForMasterSSHSocket(client.ClientId, time.Second*10, 2)
	if mwerr != nil {
		sfui.RemoveClient(&client)
		return client, mwerr
	}

	FileBrowserProxy, perr := NewHttpToUnixProxy(sfui.getFileBrowserSocketPath(client.ClientId))
	if perr != nil {
		sfui.RemoveClient(&client)
		return client, perr
	}

	client.FileBrowserProxy = FileBrowserProxy

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
	sfui.workDirRemoveClient(client.ClientId)
	delete(clients, client.ClientId)
}

// If a client has no active terminals or a GUI connection
// consider them as inactive , wait for ClientInactivityTimeout
// and then tear down the master SSH connection
func (sfui *SfUI) RemoveClientIfInactive(clientSecret string) {
	go func() {
		// Obtain a fresh copy of the client
		client, err := sfui.GetClient(clientSecret)
		if err == nil {
			if client.TerminalsCount == 0 && !client.DesktopActive {
				select {
				case <-time.After(time.Minute * time.Duration(sfui.ClientInactivityTimeout)):
					// After timeout
					client.mu.Lock()
					sfui.RemoveClient(&client)
					// Once removed there is nothing left to unlock
					break
				case _, ok := <-client.ClientConn:
					// New connection from client
					if !ok {
						break
					}
				}
			}
		}
	}()
}

func (sfui *SfUI) GetExistingClientOrMakeNew(ClientSecret string, ClientIp string) (Client, error) {
	client, cerr := sfui.GetClient(ClientSecret)
	if cerr != nil {
		return sfui.NewClient(ClientSecret, ClientIp)
	}

	if !client.MasterSSHConnectionActive {
		werr := sfui.waitForMasterSSHSocket(client.ClientId, 5*time.Second, 2)
		if werr != nil {
			return client, werr
		}

		// serialize access to client, since it can be updated inbetween by NewClient()
		client.mu.Lock()
		defer client.mu.Unlock()

		// master SSH socket is now active, grab a fresh copy of the client
		client, cerr = sfui.GetClient(ClientSecret)
	}

	return client, cerr
}

func (sfui *SfUI) GetClient(ClientSecret string) (Client, error) {
	cmu.Lock()
	defer cmu.Unlock()

	client, ok := clients[getClientId(ClientSecret)] // race possible
	if ok {
		return client, nil
	}
	return client, errors.New("No such client")
}

func (sfui *SfUI) GetClientById(ClientId string) (Client, error) {
	cmu.Lock()
	defer cmu.Unlock()

	client, ok := clients[ClientId]
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
	defer client.MarkClientIfActive()

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
	defer client.MarkClientIfInactive()

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

func (client *Client) IncSharedDesktopConnCount() error {
	defer client.MarkClientIfActive()

	// mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]

	if fclient.SharedDesktopConnCount >= fclient.MaxSharedDesktopConn {
		return errors.New("max shares reached")
	}
	fclient.SharedDesktopConnCount += 1

	// update client details
	clients[client.ClientId] = fclient
	return nil
}

func (client *Client) DecSharedDesktopConnCount() {
	defer client.MarkClientIfInactive()

	// mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]

	if fclient.SharedDesktopConnCount > 0 {
		fclient.SharedDesktopConnCount -= 1
	}

	clients[client.ClientId] = fclient
}

func (client *Client) ActivateDesktop() {
	defer client.MarkClientIfActive()

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
	defer client.MarkClientIfInactive()

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

func (client *Client) ActivateDesktopSharing(viewOnly bool, SharedSecret string) (AlreadyShared bool) {
	// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()

	// get a fresh copy of client
	fclient := clients[client.ClientId]
	if !fclient.ShareDesktop {
		fclient.ShareDesktop = true
		fclient.SharedDesktopIsViewOnly = viewOnly
		fclient.SharedDesktopSecret = SharedSecret
		fclient.SharedDesktopConn = make(chan interface{})
		clients[client.ClientId] = fclient
		return false
	}
	return true // Already shared
}

func (client *Client) DeactivateDesktopSharing() {
	// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
	// first lock then read the fresh copy to prevent a dirty read
	client.mu.Lock()
	defer client.mu.Unlock()
	// get a fresh copy of client
	if client.ShareDesktop {
		fclient := clients[client.ClientId]
		close(fclient.SharedDesktopConn)
		fclient.ShareDesktop = false
		clients[client.ClientId] = fclient
	}
}

// If no active client connection exist, open the ClientConn channel
func (client *Client) MarkClientIfInactive() {
	client.mu.Lock()
	defer client.mu.Unlock()

	fclient := clients[client.ClientId]

	if fclient.TerminalsCount == 0 && !fclient.DesktopActive {
		if fclient.ClientActive {
			fclient.ClientActive = false
			fclient.ClientConn = make(chan interface{})
			clients[client.ClientId] = fclient
		}
	}
}

// If no active client connection exist, close the ClientConn channel
func (client *Client) MarkClientIfActive() {
	client.mu.Lock()
	defer client.mu.Unlock()

	fclient := clients[client.ClientId]

	if fclient.TerminalsCount > 0 || fclient.DesktopActive {
		if !fclient.ClientActive {
			fclient.ClientActive = true
			close(fclient.ClientConn)
			clients[client.ClientId] = fclient
		}
	}
}

func (client *Client) SetTabId(TabId string) {
	client.mu.Lock()
	defer client.mu.Unlock()

	fclient := clients[client.ClientId]
	fclient.TabId = TabId
	clients[client.ClientId] = fclient
}

// Stop New Clients from obtaining service
func (sfui *SfUI) DisableClientAccess() {
	cmu.Lock()
	AcceptClients = false
	cmu.Unlock()
}

// Disable client access before calling this function
func (sfui *SfUI) RemoveAllClients() {
	for cid := range clients {
		client := clients[cid]
		sfui.RemoveClient(&client)
	}
}
