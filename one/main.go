package main

import (
	. "autosetip"
	"crypto/aes"
	"crypto/cipher"
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

//go:embed static/config.yaml.enc
var configBytes []byte

func margin(origData []byte) []byte {
	length := len(origData)
	paddingLen := int(origData[length-1])
	return origData[:(length - paddingLen)]
}

func aesDecrypt(key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(configBytes))
	blockMode.CryptBlocks(origData, configBytes)
	origData = margin(origData)
	return origData, nil
}

func readConfig(content []byte) (Config, error) {
	var config Config
	err := yaml.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}
	if len(config.IpApiURL) == 0 {
		defaultApi := []string{"https://ips.im/api", "https://api.ipify.org"}
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

func getPass() string {
	if len(os.Args) < 2 {
		return ""
	}
	return os.Args[1]
}

func main() {
	pass := getPass()
	decrypt, err := aesDecrypt([]byte(pass))
	if err != nil {
		panic(err)
	}
	config, err := readConfig(decrypt)
	if err != nil {
		fmt.Printf("Read config error: %v\n", err)
		return
	}
	Autosetip(config)
}
