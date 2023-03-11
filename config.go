package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v2"
)

func (sfui *SfUI) ReadConfig() {
	// Any options not present in config.yaml will have default values
	sfui.loadDefaultConfig()
	sfui.compileClientConfig()

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Println("Failed to read file : ", err, ", Using default configs")
		out, err := yaml.Marshal(sfui)
		if err == nil {
			os.WriteFile("config.yaml", out, os.ModeAppend)
		}
		return
	}

	var config = SfUI{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Println("Failed Unmarshal data", err)
		return
	}
	*sfui = config
}

func (sfui *SfUI) loadDefaultConfig() {
	*sfui = SfUI{
		MaxWsTerminals:          10,
		ServerBindAddress:       "127.0.0.1:7171",
		XpraWSAddress:           "ws://127.0.0.1:2000/",
		Debug:                   false,
		ShellCommand:            "bash",
		AddSfUIArgs:             false,
		SfUIOrigin:              "http://127.0.0.1:7171",
		DisableOriginCheck:      true,
		validSecretRegexMatcher: regexp.MustCompile(`^[a-zA-Z]+$`).MatchString,
	}
}

// Add any UI related configuration that has to be sent to client
// Store it byte format, to prevent json marshalling on every request
// See handleUIConfig()
func (sfui *SfUI) compileClientConfig() {
	compConfig := fmt.Sprintf(
		`{"max_terminals":"%d","auto_login":%s}`,
		sfui.MaxWsTerminals,
		strconv.FormatBool(!sfui.AddSfUIArgs), // Redirect client directly to dashboard if not in global mode.
	)
	sfui.CompiledClientConfig = []byte(compConfig)
}
