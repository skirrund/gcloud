package padding

import (
	"bytes"
)

type Padding interface {
	Padding(ciphertext []byte, blockSize int) []byte
	UnPadding(origData []byte) []byte
}

type pkcs struct{}

var PKCS = pkcs{}

func (pkcs) Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (pkcs) UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// func PKCS7Padding(src []byte, blockSize int) []byte {
// 	padding := blockSize - len(src)%blockSize
// 	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
// 	return append(src, padtext...)
// }

// func PKCS7UnPadding(src []byte, blockSize int) ([]byte, error) {
// 	length := len(src)
// 	unpadding := int(src[length-1])
// 	return src[:(length - unpadding)], nil
// }
