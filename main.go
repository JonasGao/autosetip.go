package main

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
)

type AliyunTarget struct {
	Name            string `yaml:"name"`
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

type AliyunClient struct {
	client  *ecs20140526.Client
	target  AliyunTarget
	options *util.RuntimeOptions
}

const descTemplate = "Auto create by autosetip.go. %s."

func log(client AliyunClient, msg string, a ...any) {
	fmt.Printf("[%s] %s\n", client.target.SecurityGroupId, fmt.Sprintf(msg, a))
}

func logErr(msg string, client AliyunClient, err error) {
	fmt.Printf("[%s] %s: %v\n", client.target.SecurityGroupId, msg, err)
}

func createClient(target AliyunTarget) (client AliyunClient, err error) {
	c := &openapi.Config{
		AccessKeyId:     tea.String(target.AccessKey),
		AccessKeySecret: tea.String(target.SecretKey),
		RegionId:        tea.String(target.Region),
	}
	c.Endpoint = tea.String(target.Endpoint)
	result, err := ecs20140526.NewClient(c)
	if err != nil {
		return client, err
	}
	return AliyunClient{target: target, client: result, options: &util.RuntimeOptions{}}, nil
}

func (aliyunClient AliyunClient) addIp(ip string, desc string) error {
	var err error
	permissions := &ecs20140526.AuthorizeSecurityGroupRequestPermissions{
		Policy:       tea.String("accept"),
		Priority:     tea.String("100"),
		IpProtocol:   tea.String("tcp"),
		SourceCidrIp: tea.String(ip),
		PortRange:    tea.String("22/22"),
		Description:  tea.String(desc),
	}
	req := &ecs20140526.AuthorizeSecurityGroupRequest{
		RegionId:        tea.String(aliyunClient.target.Region),
		SecurityGroupId: tea.String(aliyunClient.target.SecurityGroupId),
		Permissions:     []*ecs20140526.AuthorizeSecurityGroupRequestPermissions{permissions},
	}
	err = func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_, err = aliyunClient.client.AuthorizeSecurityGroupWithOptions(req, aliyunClient.options)
		if err != nil {
			return err
		}
		log(aliyunClient, "Success add ip rule")
		return nil
	}()
	return err
}

func (aliyunClient AliyunClient) modifyIp(id *string, ip string) error {
	var err error
	req := &ecs20140526.ModifySecurityGroupRuleRequest{
		RegionId:            tea.String(aliyunClient.target.Region),
		SecurityGroupId:     tea.String(aliyunClient.target.SecurityGroupId),
		SecurityGroupRuleId: id,
		Policy:              tea.String("accept"),
		Priority:            tea.String("100"),
		IpProtocol:          tea.String("tcp"),
		SourceCidrIp:        tea.String(ip),
		PortRange:           tea.String("22/22"),
	}
	err = func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		_, err = aliyunClient.client.ModifySecurityGroupRuleWithOptions(req, aliyunClient.options)
		if err != nil {
			return err
		}
		log(aliyunClient, "Success set ip")
		return nil
	}()
	return err
}

func (aliyunClient AliyunClient) queryRuleId(desc string) (*string, error) {
	req := &ecs20140526.DescribeSecurityGroupAttributeRequest{
		RegionId:        tea.String(aliyunClient.target.Region),
		SecurityGroupId: tea.String(aliyunClient.target.SecurityGroupId),
	}
	res, err := func() (res *ecs20140526.DescribeSecurityGroupAttributeResponse, err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		res, err = aliyunClient.client.DescribeSecurityGroupAttributeWithOptions(req, aliyunClient.options)
		if err != nil {
			return nil, err
		}
		return res, nil
	}()
	if err != nil {
		return nil, err
	}
	for _, permission := range res.Body.Permissions.Permission {
		if tea.StringValue(permission.Description) == desc {
			return permission.SecurityGroupRuleId, nil
		}
	}
	return nil, nil
}

func setIp(client AliyunClient, ip string) error {
	desc := fmt.Sprintf(descTemplate, client.target.Name)
	id, err := client.queryRuleId(desc)
	if err != nil {
		return err
	}
	if id == nil {
		log(client, "Not found rule id, will add new ip rule.")
		return client.addIp(ip, desc)
	}
	log(client, "Found rule, modify ip.")
	return client.modifyIp(id, ip)
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
		client, err := createClient(target)
		if err != nil {
			logErr("Create client error", client, err)
			return
		}
		err = setIp(client, ip)
		if err != nil {
			logErr("Modify ip error", client, err)
			return
		}
	}
	fmt.Println("Done")
}
