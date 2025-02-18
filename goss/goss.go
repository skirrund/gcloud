package goss

import (
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/skirrund/gcloud/utils"
)

type OssClient interface {
	NewDefaultClient() (OssClient, error)
	GetNativeWithPrefixUrl(fileName string) string
	GetNativePrefix() string
	GetFullUrlWithSign(fileName string, expiredInSec int64) (string, error)
	GetFullUrl(fileName string) string
	Upload(key string, reader io.Reader, isPrivate bool) (fileName string, err error)
	UploadFromUrl(urlStr string, isPrivate bool) (string, error)
	UploadOverwrite(key string, reader io.Reader, isPrivate bool) (fileName string, err error)
	UploadFile(fileName string, file *os.File, isPrivate bool) (string, error)
	GetBytes(fileName string) ([]byte, error)
	GetBase64(fileName string) (string, error)

	UploadFileBytes(fileName string, bs []byte, isPrivate bool) (string, error)
	UploadFileWithFullUrl(fileName string, file *os.File, isPrivate bool) (string, error)

	UploadFileBytesWithFullUrl(fileName string, bs []byte, isPrivate bool) (string, error)
	UploadFileBytesWithNativeFullUrl(fileName string, bs []byte, isPrivate bool) (string, error)

	UploadFileWithNativeFullUrl(fileName string, file *os.File, isPrivate bool) (string, error)

	DelObject(fileName string) (bool, error)
	IsObjectExist(key string) (bool, error)
}

var defaultOss OssClient

func SubStringBlackSlash(s string) string {
	if !strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "\\") {
		return s
	} else {
		s = s[1:]
	}
	return SubStringBlackSlash(s)
}

func GetFileName(fileName, nativePrefix, endpoint, bucketName, selfDomain string) string {
	j := -1

	if strings.HasPrefix(fileName, nativePrefix) {
		fileName = utils.SubStr(fileName, strings.Index(fileName, nativePrefix)+len(nativePrefix), -1)
	} else {
		if strings.HasPrefix(fileName, "http://") || strings.HasPrefix(fileName, "https://") {
			fileName = utils.SubStr(fileName, strings.Index(fileName, "://")+3, -1)
		}
		if !strings.HasPrefix(fileName, bucketName+"."+endpoint) && !strings.HasPrefix(fileName, selfDomain) {
			return fileName
		} else {
			j = strings.Index(fileName, "/")
			if j > -1 {
				fileName = utils.SubStr(fileName, j+1, -1)
			}
		}
	}
	i := strings.Index(fileName, "?")
	if i > -1 {
		fileName = utils.SubStr(fileName, 0, i)
	}
	if strings.Contains(fileName, "%") {
		fileName, _ = url.QueryUnescape(fileName)
	}
	return SubStringBlackSlash(fileName)
}

func GetDefault(c OssClient) (OssClient, error) {
	if defaultOss != nil {
		return defaultOss, nil
	}
	return NewDefault(c)
}

func NewDefault(c OssClient) (OssClient, error) {
	cli, err := c.NewDefaultClient()
	if err != nil {
		return nil, err
	}
	return cli, nil
}
