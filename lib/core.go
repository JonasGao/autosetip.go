package autosetip

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dds20151201 "github.com/alibabacloud-go/dds-20151201/v8/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"io"
	"net/http"
	"regexp"
)

type Loggable interface {
	lk() string
}

type EcsTarget struct {
	Region          string `yaml:"region"`
	AccessKey       string `yaml:"access_key"`
	SecretKey       string `yaml:"secret_key"`
	Endpoint        string `yaml:"endpoint,omitempty"`
	SecurityGroupId string `yaml:"security_group_id"`
}

type MongoTarget struct {
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	InstanceId string `yaml:"instance_id"`
}

type AliyunTarget struct {
	Name  string         `yaml:"name"`
	Ecs   []*EcsTarget   `yaml:"ecs"`
	Mongo []*MongoTarget `yaml:"mongo"`
}

type Config struct {
	IpApiURL []string        `yaml:"ip_api_url,omitempty"`
	Key      string          `yaml:"key"`
	Aliyun   []*AliyunTarget `yaml:"aliyun"`
}

type AliyunEcsClient struct {
	client   *ecs20140526.Client
	matchKey string
	target   AliyunTarget
	ecs      EcsTarget
	options  *util.RuntimeOptions
}

type AliyunMongoClient struct {
	client   *dds20151201.Client
	matchKey string
	target   AliyunTarget
	mongo    MongoTarget
	options  *util.RuntimeOptions
}

func (client AliyunEcsClient) lk() string {
	return "ecs:" + client.target.Name + ":" + client.ecs.SecurityGroupId
}

func (client AliyunMongoClient) lk() string {
	return "mongo:" + client.target.Name + ":" + client.mongo.InstanceId
}

func (client AliyunMongoClient) queryRuleId() (*string, error) {
	describeSecurityIpsRequest := &dds20151201.DescribeSecurityIpsRequest{
		DBInstanceId: tea.String(client.mongo.InstanceId),
	}
	res, err := func() (res *dds20151201.DescribeSecurityIpsResponse, err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		res, err = client.client.DescribeSecurityIpsWithOptions(describeSecurityIpsRequest, client.options)
		if err != nil {
			return nil, err
		}
		return res, nil
	}()
	if err != nil {
		return nil, err
	}
	for _, group := range res.Body.SecurityIpGroups.SecurityIpGroup {
		if tea.StringValue(group.SecurityIpGroupName) == client.matchKey {
			return group.SecurityIpList, nil
		}
	}
	return nil, nil
}

func (client AliyunMongoClient) modifyIp(ip string) error {
	modifySecurityIpsRequest := &dds20151201.ModifySecurityIpsRequest{
		DBInstanceId:        tea.String(client.mongo.InstanceId),
		SecurityIps:         tea.String(ip),
		ModifyMode:          tea.String("Cover"),
		SecurityIpGroupName: tea.String(client.matchKey),
	}
	_, err := func() (res *dds20151201.ModifySecurityIpsResponse, err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		res, err = client.client.ModifySecurityIpsWithOptions(modifySecurityIpsRequest, client.options)
		if err != nil {
			return nil, err
		}
		log(client, "Success cover ip")
		return res, nil
	}()
	return err
}

func (config *Config) init() error {
	if len(config.IpApiURL) == 0 {
		config.IpApiURL = []string{"https://ips.im/api", "https://api.ipify.org"}
	}
	for _, target := range config.Aliyun {
		for _, ecs := range target.Ecs {
			if ecs.Endpoint == "" {
				ecs.Endpoint = "ecs." + ecs.Region + ".aliyuncs.com"
			}
		}
	}
	return nil
}

const descTemplate = "Auto create by autosetip.go. For %s."

func log(target Loggable, msg string) {
	fmt.Printf("[%s] %s\n", target.lk(), msg)
}

func logErr(msg string, target Loggable, err error) {
	fmt.Printf("[%s] %s: %v\n", target.lk(), msg, err)
}

func (client AliyunEcsClient) addIp(ip string, desc string) error {
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
		RegionId:        tea.String(client.ecs.Region),
		SecurityGroupId: tea.String(client.ecs.SecurityGroupId),
		Permissions:     []*ecs20140526.AuthorizeSecurityGroupRequestPermissions{permissions},
	}
	err = func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_, err = client.client.AuthorizeSecurityGroupWithOptions(req, client.options)
		if err != nil {
			return err
		}
		log(client, "Success add ip rule")
		return nil
	}()
	return err
}

func (client AliyunEcsClient) modifyIp(id *string, ip string) error {
	var err error
	req := &ecs20140526.ModifySecurityGroupRuleRequest{
		RegionId:            tea.String(client.ecs.Region),
		SecurityGroupId:     tea.String(client.ecs.SecurityGroupId),
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
		_, err = client.client.ModifySecurityGroupRuleWithOptions(req, client.options)
		if err != nil {
			return err
		}
		log(client, "Success set ip")
		return nil
	}()
	return err
}

func (client AliyunEcsClient) queryRuleId(desc string) (*string, error) {
	req := &ecs20140526.DescribeSecurityGroupAttributeRequest{
		RegionId:        tea.String(client.ecs.Region),
		SecurityGroupId: tea.String(client.ecs.SecurityGroupId),
	}
	res, err := func() (res *ecs20140526.DescribeSecurityGroupAttributeResponse, err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		res, err = client.client.DescribeSecurityGroupAttributeWithOptions(req, client.options)
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

func setEcsSecurityIp(client AliyunEcsClient, ip string) error {
	desc := fmt.Sprintf(descTemplate, client.matchKey)
	log(client, fmt.Sprintf("Query rule: %s", desc))
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

func fetchIp(ipApi string) (string, int, error) {
	var ip string
	resp, err := http.Get(ipApi)
	if err != nil {
		return ip, 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Close response body error: %v\n", err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ip, 0, err
	}
	return string(body), resp.StatusCode, nil
}

func Autosetip(config Config) {
	err := config.init()
	if err != nil {
		fmt.Println(err)
		return
	}
	if isEmpty(config.Aliyun) {
		fmt.Println("No aliyun target")
		return
	}
	ip, done := tryFetchIp(config)
	if done {
		return
	}
	fmt.Printf("Current ip: %s\n", ip)
	for _, target := range config.Aliyun {
		if !setup(*target, ip, config) {
			return
		}
	}
	fmt.Println("Done")
}

func setMongoSecurityIp(client AliyunMongoClient, ip string) error {
	return client.modifyIp(ip)
}

func createEcsClient(target AliyunTarget, ecsTarget EcsTarget, config Config) (client AliyunEcsClient, err error) {
	c := &openapi.Config{
		AccessKeyId:     tea.String(ecsTarget.AccessKey),
		AccessKeySecret: tea.String(ecsTarget.SecretKey),
		RegionId:        tea.String(ecsTarget.Region),
	}
	c.Endpoint = tea.String(ecsTarget.Endpoint)
	result, err := ecs20140526.NewClient(c)
	if err != nil {
		return client, err
	}
	return AliyunEcsClient{
		target:   target,
		client:   result,
		ecs:      ecsTarget,
		matchKey: config.Key,
		options:  &util.RuntimeOptions{},
	}, nil
}

func createMongoClient(aliyun AliyunTarget, target MongoTarget, config Config) (client AliyunMongoClient, err error) {
	ac := &openapi.Config{
		AccessKeyId:     tea.String(target.AccessKey),
		AccessKeySecret: tea.String(target.SecretKey),
	}
	ac.Endpoint = tea.String("mongodb.aliyuncs.com")
	var c *dds20151201.Client
	c, err = dds20151201.NewClient(ac)
	if err != nil {
		return client, err
	}
	return AliyunMongoClient{
		target:   aliyun,
		client:   c,
		mongo:    target,
		matchKey: config.Key,
		options:  &util.RuntimeOptions{},
	}, nil
}

func setup(aliyun AliyunTarget, ip string, config Config) bool {
	for _, target := range aliyun.Ecs {
		client, err := createEcsClient(aliyun, *target, config)
		if err != nil {
			logErr("Create client error", client, err)
			return false
		}
		err = setEcsSecurityIp(client, ip)
		if err != nil {
			logErr("Modify ip error", client, err)
			return false
		}
	}
	for _, target := range aliyun.Mongo {
		client, err := createMongoClient(aliyun, *target, config)
		if err != nil {
			logErr("Create client error", client, err)
			return false
		}
		err = setMongoSecurityIp(client, ip)
		if err != nil {
			logErr("Modify ip error", client, err)
			return false
		}
	}
	return true
}

func matchIp(ip string) bool {
	if ip == "" {
		return false
	}
	match, err := regexp.MatchString(`^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`, ip)
	if err != nil {
		fmt.Printf("Match ip error: %v\n", err)
		return false
	}
	return match
}

func tryFetchIp(config Config) (string, bool) {
	var ip string
	var err error
	var sc int
	for _, ipApi := range config.IpApiURL {
		fmt.Printf("Fetching ip using: %s\n", ipApi)
		ip, sc, err = fetchIp(ipApi)
		if err != nil {
			fmt.Printf("Fetch ip error: %v\n", err)
		}
		fmt.Printf("Result ==> %d: %s\n", sc, ip)
		if matchIp(ip) {
			break
		}
	}
	if ip == "" {
		fmt.Println("No ip found")
		return "", true
	}
	return ip, false
}

func isEmpty(aliyun []*AliyunTarget) bool {
	if len(aliyun) == 0 {
		return true
	}
	return false
}
