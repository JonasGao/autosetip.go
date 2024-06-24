package main

import (
	. "autosetip"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

func getConfigPath() string {
	if len(os.Args) < 2 {
		return "config.yaml"
	}
	return os.Args[1]
}

func readConfig() (Config, error) {
	var config Config
	path := getConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, err
	}
	yamlData, err := os.ReadFile(path)
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
