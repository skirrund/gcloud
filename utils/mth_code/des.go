package mth_code

import (
	"bytes"
	"crypto/des"
	"encoding/base64"
	"errors"
)

const (
	DEFAULT_KEY = "1MEhD58VjFeFARU7BIbOYXNGCz5uQNp6"
)

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func MthDesEncrypt(ciphertext string) (string, error) {
	k, err := base64.RawURLEncoding.DecodeString(DEFAULT_KEY)
	if err != nil {
		return "", err
	}

	block, err := des.NewTripleDESCipher(k)
	if err != nil {
		return "", err
	}
	bs := block.BlockSize()
	src := PKCS5Padding([]byte(ciphertext), bs)
	if len(src)%bs != 0 {
		return "", errors.New("Need a multiple of the blocksize")
	}
	out := make([]byte, len(src))
	dst := out
	for len(src) > 0 {
		block.Encrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	return base64.RawURLEncoding.EncodeToString(out), nil
}

func MthDesDecrypt(ciphertext string) (string, error) {
	k, err := base64.RawURLEncoding.DecodeString(DEFAULT_KEY)
	if err != nil {
		return "", err
	}
	src, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := des.NewTripleDESCipher(k)
	if err != nil {
		return "", err
	}
	out := make([]byte, len(src))
	dst := out
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		return "", errors.New("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		block.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	out = PKCS5UnPadding(out)
	return string(out), nil
}
