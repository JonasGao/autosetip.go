package main

import (
	"fmt"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

func desc(target AliyunTarget) error {
	client, err := createClient(target)
	if err != nil {
		return err
	}
	req := &ecs20140526.DescribeSecurityGroupAttributeRequest{
		RegionId:        tea.String(target.Region),
		SecurityGroupId: tea.String(target.SecurityGroupId),
	}
	runtime := &util.RuntimeOptions{}
	err = func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		res, err := client.DescribeSecurityGroupAttributeWithOptions(req, runtime)
		if err != nil {
			return err
		}
		fmt.Printf("Result: %s", res)
		return nil
	}()
	return err
}
