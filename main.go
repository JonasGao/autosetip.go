package main

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
)

type AliyunTarget struct {
	AccessKey       string `yaml:"access_key"`
	SecretKey       string `yaml:"secret_key"`
	Region          string `yaml:"region"`
	Endpoint        string `yaml:"endpoint,omitempty"`
	SecurityGroupId string `yaml:"security_group_id"`
}

type Config struct {
	IpApiURL string         `yaml:"ip_api_url,omitempty"`
	Aliyun   []AliyunTarget `yaml:"aliyun"`
}

func createClient(target AliyunTarget) (_result *ecs20140526.Client, _err error) {
	c := &openapi.Config{
		AccessKeyId:     tea.String(target.AccessKey),
		AccessKeySecret: tea.String(target.SecretKey),
	}
	c.Endpoint = tea.String(target.Endpoint)
	_result, _err = ecs20140526.NewClient(c)
	return _result, _err
}

func modifyIp(target AliyunTarget, ip string) error {
	client, err := createClient(target)
	if err != nil {
		return err
	}
	req := &ecs20140526.ModifySecurityGroupRuleRequest{
		RegionId:        tea.String(target.Region),
		SecurityGroupId: tea.String(target.SecurityGroupId),
		Policy:          tea.String("accept"),
		Priority:        tea.String("100"),
		IpProtocol:      tea.String("tcp"),
		SourceCidrIp:    tea.String(ip),
		PortRange:       tea.String("22"),
		Description:     tea.String("Auto create by autosetip.go"),
	}
	runtime := &util.RuntimeOptions{}
	err = func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		_, err = client.ModifySecurityGroupRuleWithOptions(req, runtime)
		if err != nil {
			return err
		}

		return nil
	}()
	return err
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
	for _, target := range config.Aliyun {
		if target.Endpoint == "" {
			target.Endpoint = "ecs" + target.Region + ".aliyuncs.com"
		}
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
	if config.Aliyun == nil {
		fmt.Println("No aliyun target")
		return
	}
	ip, err := fetchIp(config)
	if err != nil {
		fmt.Printf("Fetch ip error: %v\n", err)
	}
	fmt.Printf("Current ip: %s\n", ip)
	for _, target := range config.Aliyun {
		err = modifyIp(target, ip)
		if err != nil {
			fmt.Printf("Modify ip error: %v\n", err)
		}
	}
	fmt.Println("Done")
}
