package alioss

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/goss"
	"github.com/skirrund/gcloud/logger"

	"github.com/skirrund/gcloud/utils"

	"bytes"

	"github.com/skirrund/gcloud/server/http"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

const (
	endpointInternalKey     = "alioss.endpoint.internal"
	accessKeyIdKey          = "alioss.accessKeyId"
	accessKeySecretKey      = "alioss.accessKeySecret"
	selfDomainKey           = "alioss.selfDomain"
	authVersionKey          = "alioss.authVersion" //"v1,v4等";
	regionKey               = "alioss.region"
	defaultEndpoint         = "oss-cn-shanghai.aliyuncs.com"
	defaultSelfDomainHost   = "static-core.demo.com"
	defaultEndpointInternal = "oss-cn-shanghai.aliyuncs.com"
)

type OssClient struct {
	BucketName string
	C          *oss.Client
}

// DelObject implements goss.OssClient.
func (c OssClient) DelObject(fileName string) (bool, error) {
	key := c.getFileName(fileName)
	if ex, _ := c.C.IsObjectExist(context.TODO(), c.BucketName, key); ex {
		req := &oss.DeleteObjectRequest{
			Bucket: oss.Ptr(c.BucketName),
			Key:    oss.Ptr(key),
		}
		_, err := c.C.DeleteObject(context.TODO(), req)
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
	return c.C.IsObjectExist(context.TODO(), c.BucketName, key)
}

// Upload implements goss.OssClient.
func (o OssClient) Upload(key string, reader io.Reader, isPrivate bool) (fileName string, err error) {
	return o.doUpload(key, reader, isPrivate, false)
}

func (c OssClient) doUploadWithContentType(key, contentType string, reader io.Reader, isPrivate, overwrite bool) (fileName string, err error) {
	acl := oss.ObjectACLPrivate
	if !isPrivate {
		acl = oss.ObjectACLPublicRead
	}
	key = subStringBlackSlash(key)
	req := &oss.PutObjectRequest{
		Acl:         acl,
		Bucket:      oss.Ptr(c.BucketName),
		Key:         oss.Ptr(key),
		ContentType: oss.Ptr(contentType),
	}

	if !overwrite {
		ex, err := c.C.IsObjectExist(context.TODO(), c.BucketName, key)
		if err != nil {
			logger.Error("[alioss] error:" + err.Error())
			return fileName, err
		}
		if ex {
			return key, errors.New("file already exists")
		}
	}
	req.Payload = reader
	_, err = c.C.PutObject(context.TODO(), req)
	if err != nil {
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	return key, err
}

func (c OssClient) doUpload(key string, reader io.Reader, isPrivate, overwrite bool) (fileName string, err error) {
	ct := utils.GetcontentType(key)
	return c.doUploadWithContentType(key, ct, reader, isPrivate, overwrite)
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
func (c OssClient) UploadFromUrl(urlStr string, isPrivate bool) (string, error) {
	i2 := strings.LastIndex(urlStr, "/")
	filenameExtension := utils.SubStr(urlStr, i2, -1)
	qi := strings.Index(filenameExtension, "?")
	li := strings.LastIndex(filenameExtension, ".")
	if li > -1 {
		filenameExtension = utils.Uuid() + utils.SubStr(filenameExtension, li, qi-li)
	} else {
		filenameExtension = utils.Uuid()
	}
	fileName := "downLoadFromUrl/" + filenameExtension
	var downLoad []byte
	resp, err := http.DefaultClient.GetUrl(urlStr, nil, nil, &downLoad)
	if err != nil {
		logger.Error("[alioss] download error:" + err.Error())
		return "", err
	}
	cts := resp.Headers["Content-Type"]
	if len(cts) > 0 {
		ct := cts[0]
		fileName, err = c.doUploadWithContentType(fileName, ct, bytes.NewReader(downLoad), isPrivate, false)
	} else {
		fileName, err = c.doUpload(fileName, bytes.NewReader(downLoad), isPrivate, false)
	}
	if err != nil {
		return fileName, err
	}
	return c.GetFullUrl(fileName), err
}

// UploadOverwrite implements goss.OssClient.
func (c OssClient) UploadOverwrite(key string, reader io.Reader, isPrivate bool) (fileName string, err error) {
	return c.doUpload(key, reader, isPrivate, true)
}

func (OssClient) NewClient(endpoint, accessKeyID, accessKeySecret, bucketName, region string) (c goss.OssClient, err error) {
	if len(endpoint) == 0 {
		err = errors.New("endpoint is  empty")
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	if len(accessKeyID) == 0 {
		err = errors.New("accessKeyID is  empty")
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	if len(accessKeySecret) == 0 {
		err = errors.New("accessKeySecret is  empty")
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	if len(bucketName) == 0 {
		err = errors.New("bucketName is  empty")
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	if len(region) == 0 {
		err = errors.New("region must not be blank when authVersion is v4")
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	ossCfg := oss.LoadDefaultConfig().WithInsecureSkipVerify(true).WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret)).WithEndpoint(endpoint).WithRegion(region).WithSignatureVersion(oss.SignatureVersionV4)
	cl := oss.NewClient(ossCfg)
	c = &OssClient{
		BucketName: bucketName,
		C:          cl,
	}
	return
}

func (oc OssClient) NewDefaultClient() (c goss.OssClient, err error) {
	cfg := env.GetInstance()
	endpointPublic := cfg.GetStringWithDefault(goss.AliEndpointPublicKey, "oss-cn-shanghai.aliyuncs.com")
	endpointInternal := cfg.GetStringWithDefault(endpointInternalKey, "oss-cn-shanghai.aliyuncs.com")
	accessKeyID := cfg.GetString(accessKeyIdKey)
	accessKeySecret := cfg.GetString(accessKeySecretKey)
	bucketName := cfg.GetStringWithDefault(goss.AlibucketNameKey, "mth-core")
	authVersion := cfg.GetString(authVersionKey)
	authVersion = strings.ToLower(authVersion)
	region := cfg.GetStringWithDefault(regionKey, "cn-shanghai")
	if len(endpointInternal) == 0 {
		if len(endpointPublic) == 0 {
			err = errors.New("[endpointInternal,endpointPublic] are both empty")
			logger.Error("[alioss] error:" + err.Error())
			return
		} else {
			endpointInternal = endpointPublic
		}
	}
	if len(accessKeyID) == 0 {
		err = errors.New("accessKeyID is  empty")
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	if len(accessKeySecret) == 0 {
		err = errors.New("accessKeySecret is  empty")
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	if len(bucketName) == 0 {
		err = errors.New("bucketName is  empty")
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	authv := oss.SignatureVersionV1
	if len(authVersion) > 0 {
		switch authVersion {
		case "v4":
			authv = oss.SignatureVersionV4
			if len(region) == 0 {
				err = errors.New("region must not be blank when authVersion is v4")
				logger.Error("[alioss] error:" + err.Error())
				return
			}
		}
	}
	ossCfg := oss.LoadDefaultConfig().WithInsecureSkipVerify(true).WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret)).WithEndpoint(endpointInternal).WithRegion(region).WithSignatureVersion(authv)
	cl := oss.NewClient(ossCfg)
	c = &OssClient{
		BucketName: bucketName,
		C:          cl,
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
	return cfg.GetStringWithDefault(goss.AlinativePrefixKey, goss.DefaultAliPrefix)
}

func getEndpoint() string {
	cfg := env.GetInstance()
	return cfg.GetStringWithDefault(goss.AliEndpointPublicKey, "oss-cn-shanghai.aliyuncs.com")
}

func getSelfDomain() bool {
	cfg := env.GetInstance()
	return cfg.GetBool(selfDomainKey)
}

func getSelfDomainHost() string {
	cfg := env.GetInstance()
	return cfg.GetString(goss.AliossSelfDomainHostKey)
}

func (c OssClient) GetFullUrl(fileName string) string {
	selfDomainHost := getSelfDomainHost()
	endpoint := getEndpoint()
	selfDomain := getSelfDomain()
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
	j := -1
	nativePrefix := GetNativePrefix()
	endpoint := getEndpoint()

	if strings.HasPrefix(fileName, nativePrefix) {
		fileName = utils.SubStr(fileName, strings.Index(fileName, nativePrefix)+len(nativePrefix), -1)
	} else {
		if strings.HasPrefix(fileName, "http://") || strings.HasPrefix(fileName, "https://") {
			fileName = utils.SubStr(fileName, strings.Index(fileName, "://")+3, -1)
		}
		if !strings.HasPrefix(fileName, c.BucketName+"."+endpoint) && !strings.HasPrefix(fileName, getSelfDomainHost()) {
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
	return subStringBlackSlash(fileName)
}

func (c OssClient) GetSignUrl(fileName string, expiredInSec int64) (string, error) {
	name := c.getFileName(fileName)
	if ex, _ := c.C.IsObjectExist(context.TODO(), c.BucketName, name); ex {
		params := utils.GetStringParamsMapFromUrl(fileName)
		delete(params, "Signature")
		delete(params, "Expires")
		delete(params, "OSSAccessKeyId")
		delete(params, "x-oss-credential")
		delete(params, "x-oss-date")
		delete(params, "x-oss-expires")
		delete(params, "x-oss-signature")
		delete(params, "x-oss-signature-version")
		// var options []oss.Option
		req := &oss.GetObjectRequest{
			Bucket: oss.Ptr(c.BucketName),
			Key:    oss.Ptr(name),
		}
		req.Parameters = params
		// for k, v := range params {
		// 	options = append(options, oss.AddParam(k, v))
		// }
		signUrl, err := c.C.Presign(context.TODO(), req, oss.PresignExpires(time.Duration(expiredInSec)*time.Second))
		if err == nil {
			sUrl := signUrl.URL
			index := strings.Index(sUrl, "?")
			name, _ = url.QueryUnescape(utils.SubStr(sUrl, 0, index))
			name = name + utils.SubStr(sUrl, index, -1)
		}
		return name, err
	} else {
		return fileName, errors.New("file not exists:" + name)
	}
}

func (c OssClient) GetFullUrlWithSign(fileName string, expiredInSec int64) (string, error) {
	url, err := c.GetSignUrl(fileName, expiredInSec)
	if err != nil {
		return url, err
	}
	fileName = utils.SubStr(url, strings.Index(url, "://")+3, -1)

	fileName = subStringBlackSlash(utils.SubStr(fileName, strings.Index(fileName, "/")+1, -1))
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
	fileName = c.getFileName(fileName)
	if ex, _ := c.C.IsObjectExist(context.TODO(), c.BucketName, fileName); ex {
		req := &oss.GetObjectRequest{
			Bucket: oss.Ptr(c.BucketName),
			Key:    oss.Ptr(fileName),
		}
		objResult, err := c.C.GetObject(context.TODO(), req)
		if err == nil {
			return io.ReadAll(objResult.Body)
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
