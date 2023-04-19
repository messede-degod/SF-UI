package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

//go:embed other/systemd/sfui.service
var serviceTemplate string

//go:embed config.yaml
var configTemplate string

const (
	InstallationDir  = "/opt/SFUI"
	BinaryName       = "sfui"
	ConfigFile       = "config.yaml"
	PidFile          = "/dev/shm/sfui.pid"
	VersionIndicator = "sfui.version"
)

func obtainRunLock() error {
	if SfUIRunning, pid := isRunning(); SfUIRunning { // Lock file exists
		return errors.New(fmt.Sprintf("Another SFUI process (PID : %d) is running !", pid))
	}

	// Write PID
	perr := os.WriteFile(PidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	if perr != nil {
		return perr
	}

	return nil
}

func releaseRunLock() error {
	return os.WriteFile(PidFile, []byte(fmt.Sprintf("%d", 0)), 0644)
}

func InstallService() error {
	log.Println("Installing SFUI....")

	if !isRoot() {
		return errors.New("Need root permission to install !")
	}

	if installationExists() {
		return errors.New("SFUI is already Installed !")
	}

	if SfUIrunning, pid := isRunning(); SfUIrunning {
		return errors.New(fmt.Sprintf("SFUI is running (PID: %d), please stop it first !", pid))
	}

	// set ExecStart and WorkingDirectory values in sfui.service
	serviceFile := fmt.Sprintf(serviceTemplate, InstallationDir+"/"+BinaryName, InstallationDir)
	serr := os.WriteFile("/lib/systemd/system/sfui.service", []byte(serviceFile), 0644)
	if serr != nil {
		return serr
	}

	merr := os.Mkdir(InstallationDir, 0700)
	if merr != nil {
		return merr
	}

	// sfui.version
	verr := os.WriteFile(InstallationDir+"/"+VersionIndicator, []byte(SfuiVersion), 0640)
	if verr != nil {
		return verr
	}

	// config.yaml
	cerr := os.WriteFile(InstallationDir+"/"+ConfigFile, []byte(configTemplate), 0640)
	if cerr != nil {
		return cerr
	}

	// sfui binary
	srcFile, err := os.Executable()
	if err != nil {
		return err
	}

	destFile := InstallationDir + "/" + BinaryName

	cperr := copyFile(srcFile, destFile, true)
	if cperr != nil {
		return cperr
	}

	return systemdEnableService("sfui")
}

func UnInstallService() error {
	log.Println("UnInstalling SFUI....")

	if !isRoot() {
		return errors.New("Need root permission to uninstall !")
	}

	SfUIRunning, pid := isRunning()
	if SfUIRunning {
		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}
		process.Kill()
	}

	rmerr := os.Remove("/lib/systemd/system/sfui.service")
	if rmerr != nil {
		log.Println(rmerr.Error())
	}

	rerr := os.RemoveAll(InstallationDir)
	if rerr != nil {
		log.Println(rerr.Error())
	}

	return reloadSystemdDaemon()
}

func copyFile(srcFileName string, destFileName string, replace bool) error {
	_, berr := os.Stat(destFileName)
	if berr == nil {
		os.Remove(destFileName)
	}

	srcFile, err := os.Open(srcFileName)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destFileName)
	if err != nil {
		return err
	}
	os.Chmod(destFileName, 0766)
	defer destFile.Close()

	// TODO : Check for Possible short write
	_, cerr := io.Copy(destFile, srcFile)
	if cerr != nil {
		return cerr
	}

	destFile.Sync()
	return nil
}

func installationExists() bool {
	_, serr := os.Stat("/lib/systemd/system/sfui.service")
	_, berr := os.Stat(InstallationDir + "/" + BinaryName)
	return serr == nil && berr == nil // both files exist
}

func isRunning() (isRunning bool, pid int) {
	pid, err := getSfUIPid()
	if err != nil {
		return false, 0
	}

	if isProcessSfUI(pid) {
		return true, pid
	}

	return false, 0
}

func getSfUIPid() (int, error) {
	pidBytes, err := os.ReadFile(PidFile)
	if err != nil {
		return 0, err
	}

	processPid, err := strconv.ParseInt(string(pidBytes), 10, 32)
	if err != nil {
		return 0, err
	}

	return int(processPid), nil
}

func isProcessSfUI(pid int) bool {
	procCmdLine := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdLineBytes, err := os.ReadFile(procCmdLine)
	if err != nil { // No Such Process
		return false
	}

	return strings.Contains(string(cmdLineBytes), BinaryName) // Whether process is SFUI
}

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to get current user: %s", err)
		return false
	}
	return currentUser.Username == "root"
}

func systemdEnableService(serviceName string) error {
	derr := reloadSystemdDaemon()
	if derr != nil {
		return derr
	}

	eerr := exec.Command("systemctl", "enable", serviceName).Run()
	if eerr != nil {
		return errors.New("Trouble Enabling SFUI Service ! , " + eerr.Error())
	}

	sserr := exec.Command("systemctl", "start", serviceName).Run()
	if sserr != nil {
		return errors.New("Trouble Starting SFUI Service ! , " + sserr.Error())
	}
	return nil
}

func reloadSystemdDaemon() error {
	log.Println("Reloading Systemd Daemon...")
	derr := exec.Command("systemctl", "daemon-reload").Run()
	if derr != nil {
		return errors.New("Trouble Reloading Systemd Dameon ! , " + derr.Error())
	}
	return nil
}
