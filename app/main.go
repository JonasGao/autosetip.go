package main

import (
	. "autosetip"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

func getConfigPath() string {
	if len(os.Args) == 0 {
		return "config.yaml"
	}
	return os.Args[0]
}

func readConfig() (Config, error) {
	var config Config
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		return config, err
	}
	yamlData, err := os.ReadFile(getConfigPath())
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func main() {
	config, err := readConfig()
	if err != nil {
		fmt.Printf("Read config file error: %v\n", err)
		return
	}
	Autosetip(config)
}
