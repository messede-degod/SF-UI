package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

const WORK_SUB_DIR = "/sfui" // all client dirs are held within this dir

func (sfui *SfUI) cleanWorkDir() error {
	if sfui.Debug {
		log.Println("Cleaning workdir...")
	}

	_, werr := os.Stat(sfui.WorkDirectory)
	if werr != nil {
		return errors.New("work directory doesnt exist")
	}
	// Remove sfui directory within the work directory
	os.RemoveAll(sfui.WorkDirectory + WORK_SUB_DIR)
	return os.Mkdir(sfui.WorkDirectory+WORK_SUB_DIR, os.ModePerm) // 0660 (oct) -> 384 (decimal)
}

func (sfui *SfUI) workDirAddClient(clientId string) error {
	_, werr := os.Stat(sfui.WorkDirectory + WORK_SUB_DIR)
	if werr != nil {
		return werr
	}

	if sfui.Debug {
		log.Println("Adding New workdir: ", sfui.WorkDirectory+WORK_SUB_DIR+"/"+clientId)
	}

	return os.Mkdir(sfui.WorkDirectory+WORK_SUB_DIR+"/"+clientId, os.ModePerm) // Change to more secure perms
}

func (sfui *SfUI) workDirRemoveClient(clientId string) error {
	if sfui.Debug {
		log.Println("Removing workdir: ", sfui.WorkDirectory+WORK_SUB_DIR+"/"+clientId)
	}

	return os.RemoveAll(sfui.WorkDirectory + WORK_SUB_DIR + "/" + clientId)
}

func (sfui *SfUI) prepareMasterSSHSocket(clientId string, clientSecret string, clientIp string) (*exec.Cmd, *os.File, error) {
	clientDir := sfui.WorkDirectory + WORK_SUB_DIR + "/" + clientId
	masterSSHCommand := sfui.MasterSSHCommand
	masterSSHCommand = fmt.Sprintf(masterSSHCommand, clientDir, clientDir, clientSecret, clientIp, sfui.SfEndpoint)

	cmd := exec.Command("bash", "-c", masterSSHCommand)

	mpty, ptyErr := pty.Start(cmd)

	go func() {
		if ptyErr == nil {
			cmd.Wait()
		}
		mpty.Close()
	}()

	return cmd, mpty, ptyErr
}

func (sfui *SfUI) waitForMasterSSHSocket(clientId string, sleepDuration time.Duration, tries int) error {
	clientDir := sfui.WorkDirectory + WORK_SUB_DIR + "/" + clientId
	socketPath := clientDir + "/ssh.sock"

	for tries > 0 {
		_, werr := os.Stat(socketPath)
		if werr == nil {
			return nil
		}
		tries -= 1
		time.Sleep(sleepDuration)
	}

	return errors.New("Master socket was not created in time")
}

func (sfui *SfUI) destroyMasterSSHSocket(client *Client) error {
	clientDir := sfui.WorkDirectory + WORK_SUB_DIR + "/" + client.ClientId
	destroyMasterSSHCommand := sfui.TearDownMasterSSHCommand
	destroyMasterSSHCommand = fmt.Sprintf(destroyMasterSSHCommand, clientDir, sfui.SfEndpoint)

	cmd := exec.Command("bash", "-c", destroyMasterSSHCommand)
	err := cmd.Run() // perform -O exit
	// Kill master ssh connection
	if client.MasterSSHConnectionCmd != nil { // Sometimes the connection might not have been established
		client.MasterSSHConnectionCmd.Process.Kill()
		client.MasterSSHConnectionCmd.Process.Wait()
	}
	return err
}

// func (sfui *SfUI) prepareWsBridgeSocket(clientId string, clientSecret string, clientIp string) error {
// 	clientDir := sfui.WorkDirectory + WORK_SUB_DIR + "/" + clientId
// 	prepareBridgeSSHCommand := sfui.GUIBridgeCommand
// 	prepareBridgeSSHCommand = fmt.Sprintf(prepareBridgeSSHCommand, clientDir, clientDir, clientSecret, clientIp, sfui.SfEndpoint)

// 	cmd := exec.Command("bash", "-c", prepareBridgeSSHCommand)
// 	go cmd.Run()
// 	return nil
// }

// Provide a SSH connection command with respect to a clients master SSH socket
func (sfui *SfUI) getSlaveSSHTerminalCommand(clientId string, clientSecret string, clientIp string) *exec.Cmd {
	clientDir := sfui.WorkDirectory + WORK_SUB_DIR + "/" + clientId
	slaveSSHTerminalCommand := sfui.SlaveSSHCommand
	slaveSSHTerminalCommand = fmt.Sprintf(slaveSSHTerminalCommand, clientDir, clientSecret, clientIp, sfui.SfEndpoint)

	return exec.Command("bash", "-c", slaveSSHTerminalCommand)
}

func (sfui *SfUI) getGUISocketPath(clientId string) string {
	return sfui.WorkDirectory + WORK_SUB_DIR + "/" + clientId + "/gui.sock"
}
