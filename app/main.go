package main

import (
	. "autosetip"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

func getArgConfigPath() string {
	if len(os.Args) < 2 {
		return "config.yaml"
	}
	return os.Args[1]
}

func getArgPass() string {
	if len(os.Args) < 3 {
		return ""
	}
	return os.Args[2]
}

func getArgMatchKey() string {
	if len(os.Args) < 4 {
		return ""
	}
	return os.Args[3]
}

func margin(origData []byte) []byte {
	length := len(origData)
	paddingLen := int(origData[length-1])
	return origData[:(length - paddingLen)]
}

func aesDecrypt(key []byte, content []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(content))
	blockMode.CryptBlocks(origData, content)
	origData = margin(origData)
	return origData, nil
}

func readConfig() (Config, error) {
	var config Config
	path := getArgConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, err
	}
	yamlBytes, err := os.ReadFile(path)
	pass := getArgPass()
	if pass != "" {
		yamlBytes, err = aesDecrypt([]byte(pass), yamlBytes)
	}
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlBytes, &config)
	if err != nil {
		return config, err
	}
	matchKey := getArgMatchKey()
	if matchKey != "" {
		config.Key = matchKey
		fmt.Printf("Override match key: %s\n", matchKey)
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
