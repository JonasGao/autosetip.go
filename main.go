package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
)

type AliyunTarget struct {
	AccessKey       string `yaml:"access_key"`
	SecretKey       string `yaml:"secret_key"`
	SecurityGroupId string `yaml:"security_group_id"`
}

type Config struct {
	IpApiURL string         `yaml:"ip_api_url,omitempty"`
	Aliyun   []AliyunTarget `yaml:"aliyun"`
}

func readConfig() (Config, error) {
	var config Config
	const defaultApi = "https://ips.im/api"
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
	if config.IpApiURL == "" {
		config.IpApiURL = defaultApi
	}
	return config, nil
}

func fetchIp(config Config) (string, error) {
	var ip string
	resp, err := http.Get(config.IpApiURL)
	if err != nil {
		return ip, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Close response body error: %v\n", err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ip, err
	}
	return string(body), nil
}

func main() {
	config, err := readConfig()
	if err != nil {
		fmt.Printf("Read config file error: %v\n", err)
		return
	}
	ip, err := fetchIp(config)
	if err != nil {
		fmt.Printf("Fetch ip error: %v\n", err)
	}
	fmt.Println(ip)
}
