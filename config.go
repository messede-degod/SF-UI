package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

func ReadConfig() SfUI {
	// Any options not present in config.yaml will have default values
	sfuiConfig := getDefaultConfig()

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Println("Failed to read file : ", err, ", Using default configs")
	}

	err = yaml.Unmarshal(data, &sfuiConfig)
	if err != nil {
		log.Println("Failed Unmarshal data", err)
	}

	sfuiConfig.CompiledClientConfig = getcompiledClientConfig(sfuiConfig)
	return sfuiConfig
}

func getDefaultConfig() SfUI {
	return SfUI{
		MaxWsTerminals:           10,
		ServerBindAddress:        "127.0.0.1:7171",
		Debug:                    false,
		MasterSSHCommand:         "sshpass -p segfault ssh -M -S %s/ssh.sock -L %s/gui.sock:127.0.0.1:5900 -o \"SetEnv SECRET=%s REMOTE_ADDR=%s\" root@%s -t sh",
		TearDownMasterSSHCommand: "sshpass -p segfault ssh -S %s/ssh.sock -O exit root@%s",
		SlaveSSHCommand:          "sshpass -p segfault ssh -S %s/ssh.sock -o \"SetEnv SECRET=%s REMOTE_ADDR=%s\" root@%s",
		SfEndpoint:               "teso.segfault.net",
		SfUIOrigin:               "http://127.0.0.1:7171",
		DisableOriginCheck:       true,
		UseXForwardedForHeader:   false,
		DisableDesktop:           false,
		WorkDirectory:            "/dev/shm/",
		StartXpraCommand:         "[[ $(ss -lnt) == *2000* ]] || /sf/bin/startxweb \n",
		StartVNCCommand:          "[[ $(ss -lnt) == *5900* ]] || /sf/bin/startxvnc \n",
		StartFileBrowserCommand:  "[[ $(ss -lnt) == *2900* ]] || /sf/bin/startfb \n",
	}
}

func getcompiledClientConfig(sfui SfUI) []byte {
	// Add any UI related configuration that has to be sent to client
	// Store it byte format, to prevent json marshalling on every request
	// See handleUIConfig()
	compConfig := []byte(fmt.Sprintf(
		`{"max_terminals":"%d","sf_endpoint":"%s","desktop_disabled":%s}`,
		sfui.MaxWsTerminals,
		sfui.SfEndpoint,
		strconv.FormatBool(sfui.DisableDesktop), // Hide the GUI Option in UI
	))
	return compConfig
}
