package zijieoss

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/goss"
	"github.com/skirrund/gcloud/logger"

	"github.com/skirrund/gcloud/utils"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

const (
	endpointInternalKey     = "zijie.oss.endpoint.internal"
	accessKeyKey            = "zijie.oss.accessKeyId"
	accessKeySecretKey      = "zijie.oss.accessKeySecret"
	selfDomainKey           = "zijie.oss.selfDomain"
	authVersionKey          = "zijie.oss.authVersion"
	regionKey               = "zijie.oss.region"
	defaultEndpoint         = "tos-cn-shanghai.volces.com"
	defaultSelfDomainHost   = "static-core.demo.com"
	defaultEndpointInternal = defaultEndpoint
	defaultRegion           = "cn-shanghai"
	defaultBucketName       = "gcloud-core"
)

type ZijieOssClient struct {
	ossClient  *tos.ClientV2
	bucketName string
}

// var ContentTypes = make(map[string]string)

// func init() {
// 	ContentTypes[".bmp"] = "image/bmp"
// 	ContentTypes[".tif"] = "image/tiff"
// 	ContentTypes[".tiff"] = "image/tiff"
// 	ContentTypes[".gif"] = "image/gif"
// 	ContentTypes[".jpeg"] = "image/jpeg"
// 	ContentTypes[".png"] = "image/png"
// 	ContentTypes[".jpg"] = "image/jpeg"
// 	ContentTypes[".html"] = "text/html"
// 	ContentTypes[".htm"] = "text/html"
// 	ContentTypes[".txt"] = "text/plain"
// 	ContentTypes[".pdf"] = "application/pdf"
// 	ContentTypes[".vsd"] = "application/vnd.visio"
// 	ContentTypes[".pptx"] = "application/vnd.ms-powerpoint"
// 	ContentTypes[".ppt"] = "application/vnd.ms-powerpoint"
// 	ContentTypes[".docx"] = "application/msword"
// 	ContentTypes[".doc"] = "application/msword"
// 	ContentTypes[".xls"] = "application/vnd.ms-excel"
// 	ContentTypes[".xlsx"] = "application/vnd.ms-excel"
// 	ContentTypes[".apk"] = "application/vnd.android.package-archive"
// 	ContentTypes[".ipa"] = "application/vnd.iphone"
// 	ContentTypes[".xml"] = "text/xml"
// 	ContentTypes[".mp3"] = "audio/mp3"
// 	ContentTypes[".wav"] = "audio/wav"
// 	ContentTypes[".au"] = "audio/basic"
// 	ContentTypes[".m3u"] = "audio/mpegurl"
// 	ContentTypes[".mid"] = "audio/mid"
// 	ContentTypes[".midi"] = "audio/mid"
// 	ContentTypes[".rmi"] = "audio/mid"
// 	ContentTypes[".wma"] = "audio/x-ms-wma"
// 	ContentTypes[".mpga"] = "audio/rn-mpeg"
// 	ContentTypes[".rmvb"] = "application/vnd.rn-realmedia-vbr"
// 	ContentTypes[".mp4"] = "video/mp4"
// 	ContentTypes[".avi"] = "video/avi"
// 	ContentTypes[".movie"] = "video/x-sgi-movie"
// 	ContentTypes[".mpa"] = "video/x-mpg"
// 	ContentTypes[".mpeg"] = "video/mpg"
// 	ContentTypes[".mpg"] = "video/mpg"
// 	ContentTypes[".mpv"] = "video/mpg"
// 	ContentTypes[".wm"] = "video/x-ms-wm"
// 	ContentTypes[".wmv"] = "video/x-ms-wmv"
// }

func (ZijieOssClient) NewClient(endpoint, ak, sk, bucketName, region, selfDomain string) (c goss.OssClient, err error) {
	if len(endpoint) == 0 {
		err = errors.New("endpoint is  empty")
		logger.Error("[zijieoss] error:" + err.Error())
		return
	}
	if len(ak) == 0 {
		err = errors.New("ak is  empty")
		logger.Error("[zijieoss] error:" + err.Error())
		return
	}
	if len(sk) == 0 {
		err = errors.New("sk is  empty")
		logger.Error("[zijieoss] error:" + err.Error())
		return
	}
	if len(bucketName) == 0 {
		err = errors.New("bucketName is  empty")
		logger.Error("[zijieoss] error:" + err.Error())
		return
	}
	if len(region) == 0 {
		err = errors.New("region must not be blank when authVersion is v4")
		logger.Error("[zijieoss] error:" + err.Error())
		return
	}
	client, err := tos.NewClientV2(endpoint, tos.WithRegion(region), tos.WithCredentials(tos.NewStaticCredentials(ak, sk)))
	if err != nil {
		logger.Error("[zijieoss] error:" + err.Error())
		return nil, err
	}
	return ZijieOssClient{ossClient: client, bucketName: bucketName}, nil
}

func (oc ZijieOssClient) NewDefaultClient() (c goss.OssClient, err error) {
	cfg := env.GetInstance()
	endpointPublic := cfg.GetStringWithDefault(goss.ZijieEndpointPublicKey, defaultEndpoint)
	endpointInternal := cfg.GetStringWithDefault(endpointInternalKey, defaultEndpoint)
	ak := cfg.GetString(accessKeyKey)
	sk := cfg.GetString(accessKeySecretKey)
	bucketName := cfg.GetStringWithDefault(goss.ZijieBucketNameKey, defaultBucketName)
	region := cfg.GetStringWithDefault(regionKey, defaultRegion)
	if len(endpointInternal) == 0 {
		if len(endpointPublic) == 0 {
			err = errors.New("[endpointInternal,endpointPublic] are both empty")
			logger.Error("[zijieoss] error:" + err.Error())
			return
		} else {
			endpointInternal = endpointPublic
		}
	}
	var selfDomain string
	// if getSelfDomain() {
	// 	selfDomain = getSelfDomainHost()
	// }
	return oc.NewClient(endpointInternal, ak, sk, bucketName, region, selfDomain)
}

func (oc ZijieOssClient) GetNativePrefix() string {
	cfg := env.GetInstance()
	return cfg.GetStringWithDefault(goss.ZijienativePrefixKey, goss.DefaultZijiePrefix)
}

func (co ZijieOssClient) GetNativeWithPrefixUrl(fileName string) string {
	return co.GetNativePrefix() + fileName
}

func getEndpoint() string {
	cfg := env.GetInstance()
	return cfg.GetStringWithDefault(goss.ZijieEndpointPublicKey, defaultEndpoint)
}

func getSelfDomain() bool {
	cfg := env.GetInstance()
	return cfg.GetBool(selfDomainKey)
}

func getSelfDomainHost() string {
	cfg := env.GetInstance()
	return cfg.GetString(goss.ZijieSelfDomainHostKey)
}

func (oc ZijieOssClient) GetFullUrl(fileName string) string {
	selfDomainHost := getSelfDomainHost()
	endpoint := getEndpoint()
	selfDomain := getSelfDomain()
	if len(selfDomainHost) == 0 {
		return "https://" + oc.bucketName + "." + endpoint + "/" + fileName
	}
	if selfDomain {
		return "https://" + selfDomainHost + "/" + fileName
	} else {
		return "https://" + oc.bucketName + "." + endpoint + "/" + fileName
	}
}

func (oc ZijieOssClient) getFileName(fileName string) string {
	nativePrefix := oc.GetNativePrefix()
	endpoint := getEndpoint()
	return goss.GetFileName(fileName, nativePrefix, endpoint, oc.bucketName, getSelfDomainHost())
}

func (oc ZijieOssClient) GetSignUrl(fileName string, expiredInSec int64) (string, error) {
	name := oc.getFileName(fileName)
	params := utils.GetStringParamsMapFromUrl(fileName)
	delete(params, "X-Tos-Algorithm")
	delete(params, "X-Tos-Credential")
	delete(params, "X-Tos-Date")
	delete(params, "X-Tos-Expires")
	delete(params, "X-Tos-SignedHeaders")
	delete(params, "X-Tos-Signature")
	delete(params, "X-Tos-Security-Token")
	delete(params, "X-Tos-Policy")
	preSignObj := &tos.PreSignedURLInput{
		HTTPMethod: enum.HttpMethodGet,
		Bucket:     oc.bucketName,
		Key:        name,
		Expires:    expiredInSec,
		Query:      params,
	}
	isCustomDomain := true
	sd := getSelfDomain()
	if sd {
		preSignObj.IsCustomDomain = &isCustomDomain
		preSignObj.AlternativeEndpoint = getSelfDomainHost()
	}
	output, err := oc.ossClient.PreSignedURL(preSignObj)
	if err == nil {
		signUrl := output.SignedUrl
		index := strings.Index(signUrl, "?")
		name, _ = url.QueryUnescape(utils.SubStr(signUrl, 0, index))
		name = name + utils.SubStr(signUrl, index, -1)
	}
	return name, err
}

func (oc ZijieOssClient) GetFullUrlWithSign(fileName string, expiredInSec int64) (string, error) {
	url, err := oc.GetSignUrl(fileName, expiredInSec)
	if err != nil {
		return url, err
	}
	fileName = utils.SubStr(url, strings.Index(url, "://")+3, -1)

	fileName = goss.SubStringBlackSlash(utils.SubStr(fileName, strings.Index(fileName, "/")+1, -1))
	if len(getSelfDomainHost()) == 0 {
		return "https://" + oc.bucketName + "." + getEndpoint() + "/" + fileName, err
	}
	if getSelfDomain() {
		return "https://" + getSelfDomainHost() + "/" + fileName, err
	} else {
		return "https://" + oc.bucketName + "." + getEndpoint() + "/" + fileName, err
	}
}

func (c ZijieOssClient) UploadFromUrl(urlStr string, isPrivate bool) (string, error) {
	res, err := http.Get(urlStr)
	if err != nil {
		logger.Error("[zijieoss] download error:" + err.Error())
		return "", err
	}
	defer res.Body.Close()
	fileName := "zjDownLoadFromUrl/" + utils.SubStr(urlStr, strings.LastIndex(urlStr, "/"), -1)
	fileName, err = c.doUpload(fileName, res.Body, isPrivate, false)
	if err != nil {
		return fileName, err
	}
	return c.GetFullUrl(fileName), err
}

func (oc ZijieOssClient) IsObjectExist(key string) (bool, error) {
	hv2 := &tos.HeadObjectV2Input{
		Bucket: oc.bucketName,
		Key:    key,
	}
	// 判断对象是否存在
	_, err := oc.ossClient.HeadObjectV2(context.Background(), hv2)
	if err != nil {
		if serverErr, ok := err.(*tos.TosServerError); ok {
			if serverErr.StatusCode == http.StatusNotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (oc ZijieOssClient) DelObject(fileName string) (bool, error) {
	key := oc.getFileName(fileName)
	hv2 := &tos.DeleteObjectV2Input{
		Bucket: oc.bucketName,
		Key:    key,
	}
	// 判断对象是否存在
	_, err := oc.ossClient.DeleteObjectV2(context.Background(), hv2)
	if err != nil {
		if serverErr, ok := err.(*tos.TosServerError); ok {
			if serverErr.StatusCode == http.StatusNotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (oc ZijieOssClient) UploadOverwrite(key string, reader io.Reader, isPrivate bool) (fileName string, err error) {
	return oc.doUpload(key, reader, isPrivate, false)
}

func (oc ZijieOssClient) doUpload(key string, reader io.Reader, isPrivate, overwrite bool) (fileName string, err error) {
	ct := utils.GetcontentType(key)
	key = goss.SubStringBlackSlash(key)
	ex, err := oc.IsObjectExist(key)
	if err != nil {
		return
	}
	if ex {
		return fileName, errors.New("object alreay exist")
	}
	v2Input := &tos.PutObjectV2Input{}
	v2Input.Bucket = oc.bucketName
	v2Input.ContentType = ct
	v2Input.Key = key
	v2Input.Content = reader
	acl := enum.ACLPrivate
	if !isPrivate {
		acl = enum.ACLPublicRead
	}
	v2Input.ACL = acl
	v2Input.ForbidOverwrite = !overwrite
	_, err = oc.ossClient.PutObjectV2(context.Background(), v2Input)
	if err != nil {
		logger.Error("[zijieoss] error:" + err.Error())
		return
	}
	return key, err
}

func (oc ZijieOssClient) Upload(key string, reader io.Reader, isPrivate bool) (fileName string, err error) {
	return oc.doUpload(key, reader, isPrivate, true)
}

func (c ZijieOssClient) UploadFile(fileName string, file *os.File, isPrivate bool) (string, error) {
	return c.doUpload(fileName, file, isPrivate, false)

}

func (c ZijieOssClient) GetBytes(fileName string) ([]byte, error) {
	fileName = c.getFileName(fileName)
	ex, err := c.IsObjectExist(fileName)
	if err != nil {
		return nil, err
	}
	if ex {
		goV2 := &tos.GetObjectV2Input{
			Bucket: c.bucketName,
			Key:    fileName,
			// 下载时重写响应头
			ResponseContentType:     "application/json",
			ResponseContentEncoding: "deflate",
		}
		obj, err := c.ossClient.GetObjectV2(context.Background(), goV2)
		if err != nil {
			return nil, err
		}
		defer obj.Content.Close()
		return io.ReadAll(obj.Content)
	} else {
		return nil, errors.New("file not exist")
	}
}

func (c ZijieOssClient) GetBase64(fileName string) (string, error) {
	b, err := c.GetBytes(fileName)
	if err == nil {
		return base64.StdEncoding.EncodeToString(b), err
	}
	return "", err
}

/*
*
@param fileName
@param file
@param isPrivate
@return http全路径
@throws Exception
*/
func (c ZijieOssClient) UploadFileWithFullUrl(fileName string, file *os.File, isPrivate bool) (string, error) {
	str, err := c.UploadFile(fileName, file, isPrivate)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

/**
 * /zijieoss/前缀
 *
 * @param fileName
 * @param bytes
 * @param isPrivate
 * @param forceUpload 文件名相同是否使用随机文件名进行上传
 */
func (c ZijieOssClient) UploadFileBytesWithNativeFullUrl(fileName string, bs []byte, isPrivate bool) (string, error) {
	str, err := c.Upload(fileName, bytes.NewReader(bs), isPrivate)
	if err == nil {
		str = c.GetNativePrefix() + str
	}
	return str, err
}

/**
 * /zijieoss/前缀
 *
 * @param fileName
 * @param file
 * @param isPrivate
 * @param forceUpload 文件名相同是否使用随机文件名进行上传
 */
func (c ZijieOssClient) UploadFileWithNativeFullUrl(fileName string, file *os.File, isPrivate bool) (string, error) {
	str, err := c.UploadFile(fileName, file, isPrivate)
	if err == nil {
		str = c.GetNativePrefix() + str
	}
	return str, err
}

func (c ZijieOssClient) UploadFileBytes(fileName string, bs []byte, isPrivate bool) (string, error) {
	return c.doUpload(fileName, bytes.NewReader(bs), isPrivate, false)
}

func (c ZijieOssClient) UploadFileBytesWithFullUrl(fileName string, bs []byte, isPrivate bool) (string, error) {
	str, err := c.doUpload(fileName, bytes.NewReader(bs), isPrivate, false)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}
