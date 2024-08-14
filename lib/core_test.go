package autosetip

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_init(t *testing.T) {
	config := Config{
		IpApiURL: []string{},
		Key:      "test",
		Aliyun: []*AliyunTarget{
			{
				Ecs: []*EcsTarget{
					{
						Port: []string{},
					},
				},
			},
		},
	}
	err := config.init()
	if err != nil {
		t.Errorf("init error: %v", err)
	}
	port := config.Aliyun[0].Ecs[0].Port[0]
	assert.Equal(t, "22", port, "The port should use default.")
}
