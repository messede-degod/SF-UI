package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http/httputil"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	ClientId                  string
	mu                        *sync.Mutex
	TerminalsCount            *atomic.Int32
	MasterSSHConnectionActive *atomic.Bool
	MasterSSHConnectionCmd    *exec.Cmd
	MasterSSHConnectionPty    *os.File
	DesktopActive             *atomic.Bool // Whether a active desktop ws connection exists
	MaxTerms                  int32
	MaxSharedDesktopConn      int32
	FileBrowserProxy          *httputil.ReverseProxy
	FileBrowserServiceActive  *atomic.Bool
	ShareDesktop              *atomic.Bool // Whether desktop sharing is active
	SharedDesktopIsViewOnly   *atomic.Bool
	SharedDesktopSecret       string
	SharedDesktopConn         chan interface{} // Channel when closed kills all shared desktop connections
	SharedDesktopConnCount    *atomic.Int32    // No of active connections to shared desktop
	// Channel when closed prevents master SSH connection from being killed by RemoveClientIfInactive,
	// that is unless a ClientInactivityTimeout is first reached, open channel indicated a inactive client
	// closed channel indicates a active client
	ClientConn   chan interface{}
	ClientActive *atomic.Bool // Atleast one active connection exists
	// Random value supplied by client during login , helps to identify duplicate sessions
	TabId   *string
	Deleted *atomic.Bool
}

var AcceptClients = true
var clients = make(map[string]Client) // Clients DB
var cmu = &sync.Mutex{}               // Synchronize access to the clients DB above
var randVal = RandomStr(10)           // Random str for deriving clientId, doesnt change unless sfui is restarted

// Return a new client, prepare necessary sockets
func (sfui *SfUI) NewClient(ClientSecret string, ClientIp string) (Client, error) {
	// Make and return a new client
	tabId := ""
	client := Client{
		ClientId:                  getClientId(ClientSecret),
		mu:                        &sync.Mutex{},
		TerminalsCount:            &atomic.Int32{},
		MasterSSHConnectionActive: &atomic.Bool{},
		MaxTerms:                  int32(sfui.MaxWsTerminals),
		MaxSharedDesktopConn:      int32(sfui.MaxSharedDesktopConn),
		ClientConn:                make(chan interface{}), // Initially no active connections exist
		ClientActive:              &atomic.Bool{},
		DesktopActive:             &atomic.Bool{},
		FileBrowserServiceActive:  &atomic.Bool{},
		ShareDesktop:              &atomic.Bool{},
		SharedDesktopConn:         make(chan interface{}),
		SharedDesktopIsViewOnly:   &atomic.Bool{},
		SharedDesktopConnCount:    &atomic.Int32{},
		Deleted:                   &atomic.Bool{},
		TabId:                     &tabId,
	}

	if !AcceptClients {
		return client, errors.New("not accepting new clients")
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
	client.MasterSSHConnectionActive.Store(true)

	cmu.Lock()
	clients[client.ClientId] = client
	cmu.Unlock()

	return client, nil
}

func (sfui *SfUI) RemoveClient(client *Client) {
	if client.Deleted == nil || client.ClientActive == nil {
		return
	}

	if client.Deleted.Load() { // already deleted
		return
	}

	client.Deleted.Store(true)

	if !client.ClientActive.Load() {
		close(client.ClientConn) // Stop RemoveClientIfInactive if running
	}

	sfui.destroyMasterSSHSocket(client)
	sfui.workDirRemoveClient(client.ClientId)

	cmu.Lock()
	delete(clients, client.ClientId)
	cmu.Unlock()
}

// If a client has no active terminals or a GUI connection
// consider them as inactive , wait for ClientInactivityTimeout
// and then tear down the master SSH connection
func (sfui *SfUI) RemoveClientIfInactive(clientSecret string) {
	go func() {
		// Obtain a fresh copy of the client
		client, err := sfui.GetClient(clientSecret)
		if err == nil {
			if client.TerminalsCount != nil && client.DesktopActive != nil {
				if client.TerminalsCount.Load() == 0 && !client.DesktopActive.Load() {
					select {
					case <-time.After(time.Minute * time.Duration(sfui.ClientInactivityTimeout)):
						// After timeout
						if client.mu != nil {
							sfui.RemoveClient(&client)
						}
						break
					case _, ok := <-client.ClientConn:
						// New connection from client
						if !ok {
							break
						}
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

	if client.MasterSSHConnectionActive == nil {
		return client, errors.New("client does not exist")
	}

	if !client.MasterSSHConnectionActive.Load() {
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
	return client, errors.New("no such client")
}

func (sfui *SfUI) GetClientById(ClientId string) (Client, error) {
	cmu.Lock()
	defer cmu.Unlock()

	client, ok := clients[ClientId]
	if ok {
		return client, nil
	}
	return client, errors.New("no such client")
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

	if client.TerminalsCount != nil {
		if client.TerminalsCount.Load() >= client.MaxTerms {
			return errors.New("max terminals allocated")
		}
		client.TerminalsCount.Add(1)
	}
	return nil
}

func (client *Client) DecTermCount() {
	defer client.MarkClientIfInactive()

	if client.TerminalsCount != nil {
		if client.TerminalsCount.Load() > 0 {
			client.TerminalsCount.Add(-1)
		}
	}
}

func (client *Client) IncSharedDesktopConnCount() error {
	defer client.MarkClientIfActive()

	if client.SharedDesktopConnCount != nil {
		if client.SharedDesktopConnCount.Load() >= client.MaxSharedDesktopConn {
			return errors.New("max shares reached")
		}
		client.SharedDesktopConnCount.Add(1)
	}
	return nil
}

func (client *Client) DecSharedDesktopConnCount() {
	defer client.MarkClientIfInactive()

	if client.SharedDesktopConnCount != nil {
		if client.SharedDesktopConnCount.Load() > 0 {
			client.SharedDesktopConnCount.Add(-1)
		}
	}
}

func (client *Client) ActivateDesktop() {
	defer client.MarkClientIfActive()

	if client.DesktopActive != nil {
		client.DesktopActive.Store(true)
	}
}

func (client *Client) DeActivateDesktop() {
	defer client.MarkClientIfActive()

	if client.DesktopActive != nil {
		client.DesktopActive.Store(false)
	}
}

func (client *Client) ActivateDesktopSharing(viewOnly bool, SharedSecret string) (AlreadyShared bool) {
	if client.mu != nil {
		// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
		// first lock then read the fresh copy to prevent a dirty read
		client.mu.Lock()
		defer client.mu.Unlock()

		// get a fresh copy of client
		fclient := clients[client.ClientId]
		if !fclient.ShareDesktop.Load() {
			fclient.ShareDesktop.Store(true)
			fclient.SharedDesktopIsViewOnly.Store(viewOnly)
			fclient.SharedDesktopSecret = SharedSecret
			fclient.SharedDesktopConn = make(chan interface{})
			if fclient.Deleted != nil {
				if !fclient.Deleted.Load() {
					clients[client.ClientId] = fclient
				}
			}
			return false
		}
	}
	return true // Already shared
}

func (client *Client) DeactivateDesktopSharing() {
	if client.mu != nil {
		// client is stale, but mu is a pointer, it locks the original Client entry in "clients"
		// first lock then read the fresh copy to prevent a dirty read
		client.mu.Lock()
		defer client.mu.Unlock()
		// get a fresh copy of client
		if client.ShareDesktop.Load() {
			fclient := clients[client.ClientId]
			close(fclient.SharedDesktopConn)
			fclient.ShareDesktop.Store(false)
			if fclient.Deleted != nil {
				if !fclient.Deleted.Load() {
					clients[client.ClientId] = fclient
				}
			}
		}
	}
}

// If no active client connection exist, open the ClientConn channel
func (client *Client) MarkClientIfInactive() {
	if client.mu != nil && client.TerminalsCount != nil && client.DesktopActive != nil {
		if client.TerminalsCount.Load() == 0 && !client.DesktopActive.Load() {
			if client.ClientActive.Load() {
				client.ClientActive.Store(false)

				cmu.Lock()
				defer cmu.Unlock()

				client.mu.Lock()
				defer client.mu.Unlock()

				fclient := clients[client.ClientId]
				if fclient.Deleted != nil {
					if !fclient.Deleted.Load() {
						fclient.ClientConn = make(chan interface{})
						clients[client.ClientId] = fclient
					}
				}

			}
		}
	}
}

// If no active client connection exist, close the ClientConn channel
func (client *Client) MarkClientIfActive() {
	if client.mu != nil && client.TerminalsCount != nil && client.DesktopActive != nil {
		if client.TerminalsCount.Load() > 0 || client.DesktopActive.Load() {
			if !client.ClientActive.Load() {
				client.ClientActive.Store(true)

				cmu.Lock()
				defer cmu.Unlock()

				client.mu.Lock()
				defer client.mu.Unlock()

				fclient := clients[client.ClientId]
				if fclient.Deleted != nil {
					if !fclient.Deleted.Load() {
						close(fclient.ClientConn)
						clients[client.ClientId] = fclient
					}
				}

			}
		}
	}
}

func (client *Client) SetTabId(TabId string) {
	if client.mu != nil {
		client.mu.Lock()
		defer client.mu.Unlock()
		client.TabId = &TabId
	}
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
