package main

import (
	. "autosetip"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

func readConfig() (Config, error) {
	var config Config
	defaultApi := []string{"https://ips.im/api", "https://api.ipify.org"}
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		config.IpApiURL = defaultApi
		return config, nil
	}
	yamlData, err := os.ReadFile("config.yaml")
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return config, err
	}
	if len(config.IpApiURL) == 0 {
		config.IpApiURL = defaultApi
	}
	for _, target := range config.Aliyun {
		for _, ecs := range target.Ecs {
			if ecs.Endpoint == "" {
				ecs.Endpoint = "ecs" + ecs.Region + ".aliyuncs.com"
			}
		}
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
