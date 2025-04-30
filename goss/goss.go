package goss

import (
	"errors"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/logger"
	"github.com/skirrund/gcloud/utils"
)

type OssType string

const (
	AliOss                  OssType = "alioss"
	ZijieOss                OssType = "zijieoss"
	AliossSelfDomainHostKey         = "alioss.selfDomainHost"
	ZijieSelfDomainHostKey          = "zijie.oss.selfDomainHost"
	AliEndpointPublicKey            = "alioss.endpoint.public"
	ZijieEndpointPublicKey          = "zijie.oss.endpoint.public"
	AlinativePrefixKey              = "alioss.nativePrefix"
	ZijienativePrefixKey            = "zijie.oss.nativePrefix"
	DefaultAliPrefix                = "/alioss-core/"
	DefaultZijiePrefix              = "/zijie-core/"
	AlibucketNameKey                = "alioss.bucketName"
	ZijieBucketNameKey              = "zijie.oss.bucketName"
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

var ossClients sync.Map

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
	k := c.GetNativePrefix()
	if v, ok := ossClients.Load(k); ok && v != nil {
		logger.Info("[goss] load ossClient from cache:", k)
		return v.(OssClient), nil
	} else {
		logger.Info("[goss] create ossClient:", k)
		v, err := NewDefault(c)
		if err != nil {
			return nil, err
		}
		ossClients.Store(k, v)
		return v, nil
	}
}

func GetOssTypeByFilePath(filePath string) (OssType, error) {
	cfg := env.GetInstance()
	prefxa := cfg.GetStringWithDefault(AlinativePrefixKey, DefaultAliPrefix)
	if strings.HasPrefix(filePath, prefxa) {
		return AliOss, nil
	}
	prefxz := cfg.GetStringWithDefault(ZijienativePrefixKey, DefaultZijiePrefix)
	if strings.HasPrefix(filePath, prefxz) {
		return ZijieOss, nil
	}
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		idx := strings.Index(filePath, "://")
		filePath = filePath[idx : idx+3]
	}
	selfDomainHosta := cfg.GetString(AliossSelfDomainHostKey)
	if len(selfDomainHosta) > 0 && strings.HasPrefix(filePath, selfDomainHosta) {
		return AliOss, nil
	}
	selfDomainHostz := cfg.GetString(ZijieSelfDomainHostKey)
	if len(selfDomainHostz) > 0 && strings.HasPrefix(filePath, selfDomainHostz) {
		return ZijieOss, nil
	}
	bucketNamea := cfg.GetString(AlibucketNameKey)
	endpointa := cfg.GetString(AliEndpointPublicKey)
	if strings.HasPrefix(filePath, bucketNamea+"."+endpointa) {
		return AliOss, nil
	}
	bucketNamez := cfg.GetString(ZijieBucketNameKey)
	endpointz := cfg.GetString(ZijieEndpointPublicKey)
	if strings.HasPrefix(filePath, bucketNamez+"."+endpointz) {
		return ZijieOss, nil
	}
	return "", errors.New("oss type not found")
}

func NewDefault(c OssClient) (OssClient, error) {
	cli, err := c.NewDefaultClient()
	if err != nil {
		return nil, err
	}
	return cli, nil
}
