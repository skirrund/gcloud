package baiduoss

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/goss"
	"github.com/skirrund/gcloud/logger"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"github.com/skirrund/gcloud/server/http"
	"github.com/skirrund/gcloud/utils"
)

const (
	endpointInternalKey     = "bdoss.endpoint.internal"
	accessKeyIdKey          = "bdoss.accessKeyId"
	accessKeySecretKey      = "bdoss.accessKeySecret"
	selfDomainKey           = "bdoss.selfDomain"
	authVersionKey          = "bdoss.authVersion" //"v1,v4等";
	defaultEndpoint         = "bj.bcebos.com"
	defaultSelfDomainHost   = "static-core.demo.com"
	defaultEndpointInternal = "bj.bcebos.com"
	cannedAclPrivate        = "private"
	cannedAclPR             = "public-read"
)

type OssClient struct {
	BucketName string
	C          *bos.Client
}

// DelObject implements goss.OssClient.
func (c OssClient) DelObject(fileName string) (bool, error) {
	key := c.getFileName(fileName)
	if ex, _ := c.IsObjectExist(fileName); ex {
		err := c.C.DeleteObject(c.BucketName, key)
		return err == nil, err
	} else {
		return false, errors.New("file not exist")
	}
}

// GetNativePrefix implements goss.OssClient.
func (o OssClient) GetNativePrefix() string {
	return GetNativePrefix()
}

func (co OssClient) GetNativeWithPrefixUrl(fileName string) string {
	return co.GetNativePrefix() + fileName
}

// IsObjectExist implements goss.OssClient.
func (c OssClient) IsObjectExist(key string) (bool, error) {
	key = c.getFileName(key)
	_, err := c.C.GetObjectMeta(c.BucketName, key)
	if err != nil {
		if realErr, ok := err.(*bce.BceServiceError); ok {
			if realErr.StatusCode == 404 {
				return false, nil
			} else {
				return false, realErr
			}
		} else {
			return false, realErr
		}
	} else {
		return true, nil
	}
}

// Upload implements goss.OssClient.
func (o OssClient) Upload(key string, reader io.Reader, isPrivate bool) (fileName string, err error) {
	return o.doUpload(key, reader, isPrivate, false)
}

func (c OssClient) doUpload(key string, reader io.Reader, isPrivate, overwrite bool) (fileName string, err error) {
	uploadReq := &goss.UploadReq{
		FileName:  key,
		Reader:    reader,
		IsPrivate: isPrivate,
		Overwrite: overwrite,
	}
	return c.doUploadByReq(uploadReq)
}

func (c OssClient) doUploadByReq(uploadReq *goss.UploadReq) (fileName string, err error) {
	key := uploadReq.FileName
	isPrivate := uploadReq.IsPrivate
	ct := uploadReq.ContentType
	if len(ct) == 0 {
		ct = utils.GetcontentType(key)
	}
	acl := cannedAclPrivate
	if !isPrivate {
		acl = cannedAclPR
	}
	key = subStringBlackSlash(key)
	overwrite := uploadReq.Overwrite
	if !overwrite {
		ex, err := c.IsObjectExist(key)
		if err != nil {
			logger.Error("[bdoss] error:" + err.Error())
			return fileName, err
		}
		if ex {
			return key, errors.New("file already exists")
		}
	}
	putObjArgs := &api.PutObjectArgs{CannedAcl: acl, ContentType: ct}
	if len(uploadReq.StorageType) > 0 {
		putObjArgs.StorageClass = getStorageType(uploadReq.StorageType)
	}
	_, err = c.C.PutObjectFromStream(c.BucketName, key, uploadReq.Reader, putObjArgs)
	if err != nil {
		logger.Error("[bdoss] error:" + err.Error())
		return
	}
	return key, err

}

// 指定Object的存储类型，
// STANDARD_IA代表低频存储，
// COLD代表冷存储，ARCHIVE代表归档存储，
// 不指定时默认是STANDARD标准存储类型；
// 如果是多AZ类型bucket，
// MAZ_STANDARD_IA代表多AZ低频存储，
// 不指定时默认是MAZ_STANDARD多AZ标准存储类型，不能是其它取值
func getStorageType(st goss.StorageType) string {
	switch st {
	case goss.StorageTypeStandard:
		return api.STORAGE_CLASS_STANDARD
	case goss.StorageTypeIA:
		return api.STORAGE_CLASS_STANDARD_IA
	case goss.StorageTypeArchive:
		return api.STORAGE_CLASS_ARCHIVE
	case goss.StorageTypeColdArchive:
		return api.STORAGE_CLASS_COLD
	default:
		return api.STORAGE_CLASS_STANDARD
	}
}

// UploadFile implements goss.OssClient.
func (c OssClient) UploadFile(fileName string, file *os.File, isPrivate bool) (string, error) {
	return c.doUpload(fileName, file, isPrivate, false)
}

// UploadFileBytes implements goss.OssClient.
func (c OssClient) UploadFileBytes(fileName string, bs []byte, isPrivate bool) (string, error) {
	return c.doUpload(fileName, bytes.NewReader(bs), isPrivate, false)
}

// UploadFileBytesWithFullUrl implements goss.OssClient.
func (c OssClient) UploadFileBytesWithFullUrl(fileName string, bs []byte, isPrivate bool) (string, error) {
	str, err := c.doUpload(fileName, bytes.NewReader(bs), isPrivate, false)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

// UploadFileBytesWithNativeFullUrl implements goss.OssClient.
func (c OssClient) UploadFileBytesWithNativeFullUrl(fileName string, bs []byte, isPrivate bool) (string, error) {
	str, err := c.doUpload(fileName, bytes.NewReader(bs), isPrivate, false)
	if err == nil {
		str = GetNativeWithPrefixUrl(str)
	}
	return str, err
}

// UploadFileWithFullUrl implements goss.OssClient.
func (c OssClient) UploadFileWithFullUrl(fileName string, file *os.File, isPrivate bool) (string, error) {
	str, err := c.doUpload(fileName, file, isPrivate, false)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

// UploadFileWithNativeFullUrl implements goss.OssClient.
func (c OssClient) UploadFileWithNativeFullUrl(fileName string, file *os.File, isPrivate bool) (string, error) {
	str, err := c.doUpload(fileName, file, isPrivate, false)
	if err == nil {
		str = GetNativeWithPrefixUrl(str)
	}
	return str, err
}

// UploadFromUrl implements goss.OssClient.
func (c OssClient) UploadFromUrl(urlStr, fileName string, isPrivate bool) (string, error) {
	i2 := utils.UnicodeLastIndex(urlStr, "/")
	filenameExtension := utils.SubStr(urlStr, i2, -1)
	qi := utils.UnicodeIndex(filenameExtension, "?")
	li := utils.UnicodeLastIndex(filenameExtension, ".")
	if li > -1 {
		filenameExtension = utils.Uuid() + utils.SubStr(filenameExtension, li, qi-li)
	} else {
		filenameExtension = utils.Uuid()
	}
	if len(fileName) == 0 {
		fileName = "bdDownLoadFromUrl/" + filenameExtension
	}
	var downLoad []byte
	resp, err := http.DefaultClient.GetUrl(urlStr, nil, nil, &downLoad)
	if err != nil {
		logger.Error("[bdoss] download error:" + err.Error())
		return "", err
	}
	cts := resp.Headers["Content-Type"]
	uploadReq := &goss.UploadReq{
		FileName:  fileName,
		Reader:    bytes.NewReader(downLoad),
		IsPrivate: isPrivate,
		Overwrite: false,
	}
	if len(cts) > 0 {
		ct := cts[0]
		uploadReq.ContentType = ct
	}
	fileName, err = c.doUploadByReq(uploadReq)
	if err != nil {
		return fileName, err
	}
	return c.GetFullUrl(fileName), err
}

// UploadOverwrite implements goss.OssClient.
func (c OssClient) UploadOverwrite(key string, reader io.Reader, isPrivate bool) (fileName string, err error) {
	return c.doUpload(key, reader, isPrivate, true)
}

func (OssClient) NewClient(endpoint, accessKeyID, accessKeySecret, bucketName string) (c goss.OssClient, err error) {
	if len(endpoint) == 0 {
		err = errors.New("endpoint is  empty")
		logger.Error("[bdoss] error:" + err.Error())
		return
	}
	if len(accessKeyID) == 0 {
		err = errors.New("accessKeyID is  empty")
		logger.Error("[bdoss] error:" + err.Error())
		return
	}
	if len(accessKeySecret) == 0 {
		err = errors.New("accessKeySecret is  empty")
		logger.Error("[bdoss] error:" + err.Error())
		return
	}
	if len(bucketName) == 0 {
		err = errors.New("bucketName is  empty")
		logger.Error("[bdoss] error:" + err.Error())
		return
	}

	bosClient, err := bos.NewClient(accessKeyID, accessKeySecret, endpoint)
	if err != nil {
		return nil, err
	}
	c = &OssClient{
		BucketName: bucketName,
		C:          bosClient,
	}
	return
}

func (oc OssClient) NewDefaultClient() (c goss.OssClient, err error) {
	cfg := env.GetInstance()
	endpointPublic := cfg.GetStringWithDefault(goss.BdEndpointPublicKey, defaultEndpoint)
	endpointInternal := cfg.GetStringWithDefault(endpointInternalKey, defaultEndpoint)
	accessKeyID := cfg.GetString(accessKeyIdKey)
	accessKeySecret := cfg.GetString(accessKeySecretKey)
	bucketName := cfg.GetStringWithDefault(goss.BdBucketNameKey, "pbm-core")
	if len(endpointInternal) == 0 {
		if len(endpointPublic) == 0 {
			err = errors.New("[endpointInternal,endpointPublic] are both empty")
			logger.Error("[bdoss] error:" + err.Error())
			return
		} else {
			endpointInternal = endpointPublic
		}
	}
	if len(accessKeyID) == 0 {
		err = errors.New("accessKeyID is  empty")
		logger.Error("[bdoss] error:" + err.Error())
		return
	}
	if len(accessKeySecret) == 0 {
		err = errors.New("accessKeySecret is  empty")
		logger.Error("[bdoss] error:" + err.Error())
		return
	}
	if len(bucketName) == 0 {
		err = errors.New("bucketName is  empty")
		logger.Error("[bdoss] error:" + err.Error())
		return
	}
	bosClient, err := bos.NewClient(accessKeyID, accessKeySecret, endpointInternal)
	if err != nil {
		return nil, err
	}
	c = &OssClient{
		BucketName: bucketName,
		C:          bosClient,
	}
	return
}

func subStringBlackSlash(s string) string {
	if !strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "\\") {
		return s
	} else {
		s = s[1:]
	}
	return subStringBlackSlash(s)
}

func GetNativeWithPrefixUrl(fileName string) string {
	return GetNativePrefix() + fileName
}

func GetNativePrefix() string {
	cfg := env.GetInstance()
	return cfg.GetStringWithDefault(goss.BdnativePrefixKey, goss.DefaultBdPrefix)
}

func getEndpoint() string {
	cfg := env.GetInstance()
	return cfg.GetStringWithDefault(goss.BdEndpointPublicKey, defaultEndpoint)
}

func getSelfDomain() bool {
	cfg := env.GetInstance()
	return cfg.GetBool(selfDomainKey)
}

func getSelfDomainHost() string {
	cfg := env.GetInstance()
	return cfg.GetString(goss.BdSelfDomainHostKey)
}

func (c OssClient) GetFullUrl(fileName string) string {
	selfDomainHost := getSelfDomainHost()
	endpoint := getEndpoint()
	selfDomain := getSelfDomain()
	nativePrefix := c.GetNativePrefix()
	fileName = goss.GetFileName(fileName, nativePrefix, endpoint, c.BucketName, getSelfDomainHost(), false)
	if len(selfDomainHost) == 0 {
		return "https://" + c.BucketName + "." + endpoint + "/" + fileName
	}
	if selfDomain {
		return "https://" + selfDomainHost + "/" + fileName
	} else {
		return "https://" + c.BucketName + "." + endpoint + "/" + fileName
	}
}

func (c OssClient) getFileName(fileName string) string {
	nativePrefix := c.GetNativePrefix()
	endpoint := getEndpoint()
	return goss.GetFileName(fileName, nativePrefix, endpoint, c.BucketName, getSelfDomainHost(), true)
}

func (c OssClient) GetSignUrl(fileName string, expiredInSec int64) (string, error) {
	name := c.getFileName(fileName)
	ex, err := c.IsObjectExist(fileName)
	if err != nil {
		return fileName, err
	}
	if ex {
		params := utils.GetStringParamsMapFromUrl(fileName)
		delete(params, "authorization")
		delete(params, "x-bce-security-token")

		// for k, v := range params {
		// 	options = append(options, oss.AddParam(k, v))
		// }
		signUrl := c.C.GeneratePresignedUrl(c.BucketName, name, int(expiredInSec), "GET", nil, params)
		sUrl := signUrl
		index := utils.UnicodeIndex(sUrl, "?")
		name, _ = url.QueryUnescape(utils.SubStr(sUrl, 0, index))
		name = name + utils.SubStr(sUrl, index, -1)
		return name, nil
	} else {
		return fileName, errors.New("file not exists:" + name)
	}
}

func (c OssClient) GetFullUrlWithSign(fileName string, expiredInSec int64) (string, error) {
	url, err := c.GetSignUrl(fileName, expiredInSec)
	if err != nil {
		return url, err
	}
	fileName = utils.SubStr(url, utils.UnicodeIndex(url, "://")+3, -1)

	fileName = subStringBlackSlash(utils.SubStr(fileName, utils.UnicodeIndex(fileName, "/")+1, -1))
	if len(getSelfDomainHost()) == 0 {
		return "https://" + c.BucketName + "." + getEndpoint() + "/" + fileName, err
	}
	if getSelfDomain() {
		return "https://" + getSelfDomainHost() + "/" + fileName, err
	} else {
		return "https://" + c.BucketName + "." + getEndpoint() + "/" + fileName, err
	}
}
func (c OssClient) GetBytes(fileName string) ([]byte, error) {
	name := c.getFileName(fileName)
	if ex, _ := c.IsObjectExist(fileName); ex {
		objResult, err := c.C.GetObject(c.BucketName, name, nil)
		if err == nil {
			bytes, err := io.ReadAll(objResult.Body)
			objResult.Body.Close()
			return bytes, err
		}
		return nil, err
	} else {
		return nil, errors.New("file not exist")
	}
}

func (c OssClient) GetBase64(fileName string) (string, error) {
	b, err := c.GetBytes(fileName)
	if err == nil {
		return base64.StdEncoding.EncodeToString(b), err
	}
	return "", err
}

func (c OssClient) UploadByReq(req *goss.UploadReq) (string, error) {
	str, err := c.doUploadByReq(req)
	if err == nil {
		str = GetNativeWithPrefixUrl(str)
	}
	return str, err
}
