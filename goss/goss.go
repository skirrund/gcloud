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

type StorageType string

const (
	// StorageClassStandard Standard provides highly reliable, highly available,
	// and high-performance object storage for data that is frequently accessed.
	StorageTypeStandard StorageType = "Standard"

	// StorageClassIA IA provides highly durable storage at lower prices compared with Standard.
	// IA has a minimum billable size of 64 KB and a minimum billable storage duration of 30 days.
	StorageTypeIA StorageType = "IA"

	// StorageClassArchive Archive provides high-durability storage at lower prices compared with Standard and IA.
	// Archive has a minimum billable size of 64 KB and a minimum billable storage duration of 60 days.
	StorageTypeArchive StorageType = "Archive"

	// StorageClassColdArchive Cold Archive provides highly durable storage at lower prices compared with Archive.
	// Cold Archive has a minimum billable size of 64 KB and a minimum billable storage duration of 180 days.
	StorageTypeColdArchive StorageType = "ColdArchive"

	// StorageClassDeepColdArchive Deep Cold Archive provides highly durable storage at lower prices compared with Cold Archive.
	// Deep Cold Archive has a minimum billable size of 64 KB and a minimum billable storage duration of 180 days.
	StorageTypeDeepColdArchive StorageType = "DeepColdArchive"
	//bytedance
	StorageClassArchiveFr          StorageType = "ARCHIVE_FR"
	StorageClassIntelligentTiering StorageType = "INTELLIGENT_TIERING"
)

type OssType string

const (
	AliOss                  OssType = "alioss"
	ZijieOss                OssType = "zijieoss"
	BdOss                   OssType = "bdoss"
	AliossSelfDomainHostKey         = "alioss.selfDomainHost"
	ZijieSelfDomainHostKey          = "zijie.oss.selfDomainHost"
	BdSelfDomainHostKey             = "bd.oss.selfDomainHost"
	AliEndpointPublicKey            = "alioss.endpoint.public"
	ZijieEndpointPublicKey          = "zijie.oss.endpoint.public"
	BdEndpointPublicKey             = "bd.oss.endpoint.public"
	AlinativePrefixKey              = "alioss.nativePrefix"
	ZijienativePrefixKey            = "zijie.oss.nativePrefix"
	BdnativePrefixKey               = "bd.oss.nativePrefix"
	DefaultAliPrefix                = "/alioss-core/"
	DefaultZijiePrefix              = "/zijie-core/"
	DefaultBdPrefix                 = "/bd-core/"
	AlibucketNameKey                = "alioss.bucketName"
	ZijieBucketNameKey              = "zijie.oss.bucketName"
	BdBucketNameKey                 = "bd.oss.bucketName"
)

type OssClient interface {
	NewDefaultClient() (OssClient, error)
	GetNativeWithPrefixUrl(fileName string) string
	GetNativePrefix() string
	GetFullUrlWithSign(fileName string, expiredInSec int64) (string, error)
	GetFullUrl(fileName string) string
	Upload(key string, reader io.Reader, isPrivate bool) (fileName string, err error)
	UploadFromUrl(urlStr, targetName string, isPrivate bool) (string, error)
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

	UploadByReq(req *UploadReq) (string, error)
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

func GetFileName(fileName, nativePrefix, endpoint, bucketName, selfDomain string, cutParams bool) string {
	j := -1
	if strings.HasPrefix(fileName, nativePrefix) {
		fileName = utils.SubStr(fileName, utils.UnicodeIndex(fileName, nativePrefix)+len(nativePrefix), -1)
	} else {
		if strings.HasPrefix(fileName, "http://") || strings.HasPrefix(fileName, "https://") {
			fileName = utils.SubStr(fileName, utils.UnicodeIndex(fileName, "://")+3, -1)
		}
		if strings.HasPrefix(fileName, bucketName+"."+endpoint) || (len(selfDomain) > 0 && strings.HasPrefix(fileName, selfDomain)) {
			j = utils.UnicodeIndex(fileName, "/")
			if j > -1 {
				fileName = utils.SubStr(fileName, j+1, -1)
			}
		}
	}
	if cutParams {
		i := utils.UnicodeIndex(fileName, "?")
		if i > -1 {
			fileName = utils.SubStr(fileName, 0, i)
		}
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
	prefxb := cfg.GetStringWithDefault(BdnativePrefixKey, DefaultBdPrefix)
	if strings.HasPrefix(filePath, prefxb) {
		return BdOss, nil
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
	selfDomainHostB := cfg.GetString(BdSelfDomainHostKey)
	if len(selfDomainHostB) > 0 && strings.HasPrefix(filePath, selfDomainHostB) {
		return BdOss, nil
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
	bucketNameB := cfg.GetString(BdBucketNameKey)
	endpointB := cfg.GetString(ZijieEndpointPublicKey)
	if strings.HasPrefix(filePath, bucketNameB+"."+endpointB) {
		return BdOss, nil
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
