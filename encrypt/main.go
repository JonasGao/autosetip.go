package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"os"
)

func padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	text := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, text...)
}

func aesEncrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	data = padding(data, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encrypted := make([]byte, len(data))
	blockMode.CryptBlocks(encrypted, data)
	return encrypted, nil
}

func getConfigPath() string {
	if len(os.Args) < 3 {
		return "config.yaml"
	}
	return os.Args[2]
}

func getPass() string {
	if len(os.Args) < 2 {
		return ""
	}
	return os.Args[1]
}

func main() {
	file, err := os.ReadFile(getConfigPath())
	if err != nil {
		panic(err)
	}
	pass := getPass()
	if pass == "" {
		panic("pass is empty")
	}
	encrypt, err := aesEncrypt(file, []byte(pass))
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("config.yaml.enc", encrypt, 0644)
	if err != nil {
		panic(err)
	}
}
