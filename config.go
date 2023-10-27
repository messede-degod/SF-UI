package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"sync/atomic"

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
	sfuiConfig.NoEndpoints = int32(len(sfuiConfig.SfEndpoints))
	return sfuiConfig
}

func getDefaultConfig() SfUI {
	return SfUI{
		MaxWsTerminals:           10,
		MaxSharedDesktopConn:     4,
		ServerBindAddress:        "127.0.0.1:7171",
		Debug:                    false,
		MasterSSHCommand:         "sshpass -p segfault ssh -M -S %s/ssh.sock -L %s/gui.sock:127.0.0.1:5900 -o \"SetEnv SECRET=%s REMOTE_ADDR=%s\" root@%s -t sh",
		TearDownMasterSSHCommand: "sshpass -p segfault ssh -S %s/ssh.sock -O exit root@%s",
		SlaveSSHCommand:          "sshpass -p segfault ssh -S %s/ssh.sock -o \"SetEnv SECRET=%s REMOTE_ADDR=%s\" root@%s",
		SfEndpoints: []string{
			"8lgm.segfault.net",
			"adm.segfault.net"},
		SfUIOrigin:              "http://127.0.0.1:7171",
		DisableOriginCheck:      true,
		UseXForwardedForHeader:  false,
		DisableDesktop:          false,
		StartXpraCommand:        "[[ $(ss -lnt) == *2000* ]] || /sf/bin/startxweb \n",
		StartVNCCommand:         "[[ $(ss -lnt) == *5900* ]] || /sf/bin/startxvnc \n",
		StartFileBrowserCommand: "[[ $(ss -lnt) == *2900* ]] || /sf/bin/startfb \n",
		ClientInactivityTimeout: 3,
		WSPingInterval:          20,
		WSTimeout:               1080, // 18 Hours
		ValidSecret:             regexp.MustCompile(`^[a-zA-Z0-9-]{6,}$`).MatchString,
		EndpointSelector:        &atomic.Int32{},
		VNCPort:                 5900,
		FileBrowserPort:         2900,
		SegfaultSSHUsername:     "root",
		SegfaultSSHPassword:     "segfault",
		SegfaultUseSSHKey:       false,
		SegfaultSSHKeyPath:      "",
		MaintenanceSecret:       RandomStr(42),
		EnableMetricLogging:     false,
		MetricLoggerQueueSize:   500,
		GeoIpDBPath:             "geo.mmdb",
	}
}

type UIConfig struct {
	MaxTerms           int      `json:"max_terminals"`
	DesktopDisabled    bool     `json:"desktop_disabled"`
	WSPingInterval     int      `json:"ws_ping_interval"`
	BuildHash          string   `json:"build_hash"`
	BuildTime          string   `json:"build_time"`
	AvailableEndpoints []string `json:"available_endpoints"`
}

func getcompiledClientConfig(sfui SfUI) []byte {
	// Add any UI related configuration that has to be sent to client
	// Store it byte format, to prevent json marshalling on every request
	// See handleUIConfig()

	config := UIConfig{
		MaxTerms:           sfui.MaxWsTerminals,
		DesktopDisabled:    sfui.DisableDesktop,
		WSPingInterval:     sfui.WSPingInterval,
		BuildHash:          buildHash,
		BuildTime:          buildTime,
		AvailableEndpoints: sfui.SfEndpoints,
	}

	configBytes, err := json.Marshal(config)
	if err == nil {
		return configBytes
	}
	return nil
}
