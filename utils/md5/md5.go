package md5

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func MD5Encode(str string) string {
	b := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", b)
}

func MD5EncodeUpper(str string) string {
	return strings.ToUpper(MD5Encode(str))
}
