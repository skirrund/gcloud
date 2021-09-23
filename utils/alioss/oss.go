package alioss

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/logger"

	"github.com/skirrund/gcloud/utils"

	"bytes"

	"github.com/skirrund/gcloud/server/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const (
	selfDomainHostKey       = "alioss.selfDomainHost"
	endpointPublicKey       = "alioss.endpoint.public"
	endpointInternalKey     = "alioss.endpoint.internal"
	accessKeyIdKey          = "alioss.accessKeyId"
	accessKeySecretKey      = "alioss.accessKeySecret"
	bucketNameKey           = "alioss.bucketName" //"mth-core";
	selfDomainKey           = "alioss.selfDomain"
	nativePrefixKey         = "alioss.nativePrefix" //"=/alioss-core/";
	defaultEndpoint         = "oss-cn-shanghai.aliyuncs.com"
	defaultSelfDomainHost   = "static-core.meditrusthealth.com"
	defaultEndpointInternal = "oss-cn-shanghai.aliyuncs.com"
)

type ossClient struct {
	C *oss.Bucket
}

var client *ossClient

var ContentTypes = make(map[string]string)

func init() {
	ContentTypes[".bmp"] = "image/bmp"
	ContentTypes[".tif"] = "image/tiff"
	ContentTypes[".tiff"] = "image/tiff"
	ContentTypes[".gif"] = "image/gif"
	ContentTypes[".jpeg"] = "image/jpg"
	ContentTypes[".png"] = "image/jpg"
	ContentTypes[".jpg"] = "image/jpg"
	ContentTypes[".html"] = "text/html"
	ContentTypes[".htm"] = "text/html"
	ContentTypes[".txt"] = "text/plain"
	ContentTypes[".pdf"] = "application/pdf"
	ContentTypes[".vsd"] = "application/vnd.visio"
	ContentTypes[".pptx"] = "application/vnd.ms-powerpoint"
	ContentTypes[".ppt"] = "application/vnd.ms-powerpoint"
	ContentTypes[".docx"] = "application/msword"
	ContentTypes[".doc"] = "application/msword"
	ContentTypes[".xls"] = "application/vnd.ms-excel"
	ContentTypes[".xlsx"] = "application/vnd.ms-excel"
	ContentTypes[".apk"] = "application/vnd.android.package-archive"
	ContentTypes[".ipa"] = "application/vnd.iphone"
	ContentTypes[".xml"] = "text/xml"
	ContentTypes[".mp3"] = "audio/mp3"
	ContentTypes[".wav"] = "audio/wav"
	ContentTypes[".au"] = "audio/basic"
	ContentTypes[".m3u"] = "audio/mpegurl"
	ContentTypes[".mid"] = "audio/mid"
	ContentTypes[".midi"] = "audio/mid"
	ContentTypes[".rmi"] = "audio/mid"
	ContentTypes[".wma"] = "audio/x-ms-wma"
	ContentTypes[".mpga"] = "audio/rn-mpeg"
	ContentTypes[".rmvb"] = "application/vnd.rn-realmedia-vbr"
	ContentTypes[".mp4"] = "video/mp4"
	ContentTypes[".avi"] = "video/avi"
	ContentTypes[".movie"] = "video/x-sgi-movie"
	ContentTypes[".mpa"] = "video/x-mpg"
	ContentTypes[".mpeg"] = "video/mpg"
	ContentTypes[".mpg"] = "video/mpg"
	ContentTypes[".mpv"] = "video/mpg"
	ContentTypes[".wm"] = "video/x-ms-wm"
	ContentTypes[".wmv"] = "video/x-ms-wmv"
}

func NewClient(endpoint string, accessKeyID string, accessKeySecret string, bucketName string) (c *ossClient, err error) {
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
	cl, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	b, err := cl.Bucket(bucketName)
	if err != nil {
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	c = &ossClient{
		C: b,
	}
	return
}

func NewDefaultClient() (c *ossClient, err error) {
	if client != nil {
		c = client
		return
	}
	cfg := env.GetInstance()
	endpointPublic := cfg.GetStringWithDefault(endpointPublicKey, "oss-cn-shanghai.aliyuncs.com")
	endpointInternal := cfg.GetStringWithDefault(endpointInternalKey, "oss-cn-shanghai.aliyuncs.com")
	accessKeyID := cfg.GetString(accessKeyIdKey)
	accessKeySecret := cfg.GetString(accessKeySecretKey)
	bucketName := cfg.GetStringWithDefault(bucketNameKey, "mth-core")
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
	cl, err := oss.New(endpointInternal, accessKeyID, accessKeySecret)
	if err != nil {
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	b, err := cl.Bucket(bucketName)
	if err != nil {
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	c = &ossClient{
		C: b,
	}
	return
}

func getcontentType(fileName string) string {
	index := strings.Index(fileName, ".")
	if index == -1 {
		return "image/jpg"
	}
	filenameExtension := utils.SubStr(fileName, strings.LastIndex(fileName, "."), -1)
	if len(filenameExtension) > 0 {
		contentType := ContentTypes[strings.ToLower(filenameExtension)]
		if len(contentType) > 0 {
			return contentType
		}
	}
	return "application/octet-stream"
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
	return cfg.GetStringWithDefault(nativePrefixKey, "/alioss-core/")
}

func getEndpoint() string {
	cfg := env.GetInstance()
	return cfg.GetStringWithDefault(endpointPublicKey, "oss-cn-shanghai.aliyuncs.com")
}

func getSelfDomain() bool {
	return true
	cfg := env.GetInstance()
	return cfg.GetBool(selfDomainKey)
}

func getSelfDomainHost() string {
	cfg := env.GetInstance()

	return cfg.GetStringWithDefault(selfDomainHostKey, "static-core.meditrusthealth.com")
}

func (c *ossClient) GetFullUrl(fileName string) string {
	selfDomainHost := getSelfDomainHost()
	endpoint := getEndpoint()
	selfDomain := getSelfDomain()
	if len(selfDomainHost) == 0 {
		return "https://" + c.C.BucketName + "." + endpoint + "/" + fileName
	}
	if selfDomain {
		return "https://" + selfDomainHost + "/" + fileName
	} else {
		return "https://" + c.C.BucketName + "." + endpoint + "/" + fileName
	}
}

func (c *ossClient) getFileName(fileName string) string {
	j := -1
	nativePrefix := GetNativePrefix()
	endpoint := getEndpoint()

	if strings.HasPrefix(fileName, nativePrefix) {
		fileName = utils.SubStr(fileName, strings.Index(fileName, nativePrefix)+len(nativePrefix), -1)
	} else {
		if strings.HasPrefix(fileName, "http://") || strings.HasPrefix(fileName, "https://") {
			fileName = utils.SubStr(fileName, strings.Index(fileName, "://")+3, -1)
		}
		if !strings.HasPrefix(fileName, c.C.BucketName+"."+endpoint) && !strings.HasPrefix(fileName, getSelfDomainHost()) {
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
	if strings.Index(fileName, "%") > -1 {
		fileName, _ = url.QueryUnescape(fileName)
	}
	return subStringBlackSlash(fileName)
}

func (c *ossClient) GetSignUrl(fileName string, expiredInSec int64) (string, error) {
	name := c.getFileName(fileName)
	if ex, _ := c.C.IsObjectExist(name); ex {
		params := utils.GetStringParamsMapFromUrl(fileName)
		delete(params, "Signature")
		delete(params, "Expires")
		delete(params, "OSSAccessKeyId")
		var options []oss.Option
		for k, v := range params {
			options = append(options, oss.AddParam(k, v))
		}
		signUrl, err := c.C.SignURL(name, oss.HTTPGet, expiredInSec, options...)
		if err == nil {
			index := strings.Index(signUrl, "?")
			name, _ = url.QueryUnescape(utils.SubStr(signUrl, 0, index))
			name = name + utils.SubStr(signUrl, index, -1)
		}
		return name, err
	} else {
		return fileName, errors.New("file not exists:" + name)
	}
}

func (c *ossClient) GetFullUrlWithSign(fileName string, expiredInSec int64) (string, error) {
	url, err := c.GetSignUrl(fileName, expiredInSec)
	if err != nil {
		return url, err
	}
	fileName = utils.SubStr(url, strings.Index(url, "://")+3, -1)

	fileName = subStringBlackSlash(utils.SubStr(fileName, strings.Index(fileName, "/")+1, -1))
	if len(getSelfDomainHost()) == 0 {
		return "https://" + c.C.BucketName + "." + getEndpoint() + "/" + fileName, err
	}
	if getSelfDomain() {
		return "https://" + getSelfDomainHost() + "/" + fileName, err
	} else {
		return "https://" + c.C.BucketName + "." + getEndpoint() + "/" + fileName, err
	}
}

func (c *ossClient) UploadFromUrl(urlStr string, isPrivate bool, forceUpload bool) (string, error) {
	var downLoad []byte
	_, err := http.GetUrl(urlStr, nil, &downLoad)
	if err != nil {
		logger.Error("[alioss] download error:" + err.Error())
		return "", err
	}
	fileName := "downLoadFromUrl/" + utils.SubStr(urlStr, strings.LastIndex(urlStr, "/"), -1)
	fileName, err = c.UploadFileBytes(fileName, downLoad, isPrivate, forceUpload)
	if err != nil {
		return fileName, err
	}
	return c.GetFullUrl(fileName), err
}

func (c *ossClient) UploadFileBytes(key string, bs []byte, isPrivate bool, forceUpload bool) (fileName string, err error) {
	acl := oss.ObjectACL(oss.ACLPrivate)
	if !isPrivate {
		acl = oss.ObjectACL(oss.ACLPublicRead)
	}
	ct := getcontentType(key)
	contentType := oss.ContentType(ct)
	key = subStringBlackSlash(key)

	ex, err := c.C.IsObjectExist(key)
	if err != nil {
		logger.Error("[alioss] error:" + err.Error())
		return fileName, err
	}
	if ex {
		if forceUpload {
			err = c.C.DeleteObject(key)
			if err != nil {
				logger.Error("[alioss] error:" + err.Error())
			}
		} else {
			return key, err
		}
	}
	err = c.C.PutObject(key, bytes.NewReader(bs), contentType, acl)
	if err != nil {
		logger.Error("[alioss] error:" + err.Error())
		return
	}
	return key, err
}

func (c *ossClient) UploadFileFile(fileName string, file *os.File, isPrivate bool, forceUpload bool) (string, error) {
	b, err := ioutil.ReadAll(file)
	if err == nil {
		return c.UploadFileBytes(fileName, b, isPrivate, forceUpload)
	}
	return "", err

}

func (c *ossClient) GetBytes(fileName string) ([]byte, error) {
	fileName = c.getFileName(fileName)
	if ex, _ := c.C.IsObjectExist(fileName); ex {
		obj, err := c.C.GetObject(fileName)
		if err == nil {
			return ioutil.ReadAll(obj)
		}
		return nil, err
	} else {
		return nil, errors.New("file not exist")
	}
}

func (c *ossClient) GetBase64(fileName string) (string, error) {
	b, err := c.GetBytes(fileName)
	if err == nil {
		return base64.StdEncoding.EncodeToString(b), err
	}
	return "", err
}

/**
@param fileName
@param file
@param isPrivate
@param forceUpload 文件名相同是否使用随机文件名进行上传
@return http全路径
@throws Exception
 */
func (c *ossClient) UploadFileWithFullUrl(fileName string, file *os.File, isPrivate bool, forceUpload bool) (string, error) {
	str, err := c.UploadFileFile(fileName, file, isPrivate, forceUpload)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

/**
@param fileName
@param bytes
@param isPrivate
@param forceUpload 文件名相同是否使用随机文件名进行上传
@return http全路径
@throws Exception
 */
func (c *ossClient) UploadFileBytesWithFullUrl(fileName string, bs []byte, isPrivate bool, forceUpload bool) (string, error) {
	str, err := c.UploadFileBytes(fileName, bs, isPrivate, forceUpload)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

/**
 * /alioss/前缀
 *
 * @param fileName
 * @param bytes
 * @param isPrivate
 * @param forceUpload 文件名相同是否使用随机文件名进行上传
 */
func (c *ossClient) UploadFileBytesWithNativeFullUrl(fileName string, bs []byte, isPrivate bool, forceUpload bool) (string, error) {
	str, err := c.UploadFileBytes(fileName, bs, isPrivate, forceUpload)
	if err == nil {
		str = GetNativeWithPrefixUrl(str)
	}
	return str, err
}

/**
 * /alioss/前缀
 *
 * @param fileName
 * @param file
 * @param isPrivate
 * @param forceUpload 文件名相同是否使用随机文件名进行上传
 */
func (c *ossClient) UploadFileWithNativeFullUrl(fileName string, file *os.File, isPrivate bool, forceUpload bool) (string, error) {
	str, err := c.UploadFileFile(fileName, file, isPrivate, forceUpload)
	if err == nil {
		str = GetNativeWithPrefixUrl(str)
	}
	return str, err
}

func (c *ossClient) UploadPublicFileBytes(fileName string, bs []byte) (string, error) {
	return c.UploadFileBytes(fileName, bs, false, true)
}

func (c *ossClient) UploadPrivateFileBytes(fileName string, bs []byte) (string, error) {
	return c.UploadFileBytes(fileName, bs, true, true)
}

func (c *ossClient) UploadPublicFileBytesWithFullUrl(fileName string, bs []byte) (string, error) {
	str, err := c.UploadFileBytes(fileName, bs, false, true)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

func (c *ossClient) UploadPrivateFileBytesWithFullUrl(fileName string, bs []byte) (string, error) {
	str, err := c.UploadFileBytes(fileName, bs, true, true)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

/**
 * /alioss/前缀
 *
 * @param fileName
 * @param bytes

 */

func (c *ossClient) UploadPrivateFileBytesWithNativeFullUrl(fileName string, bs []byte) (string, error) {

	str, err := c.UploadFileBytes(fileName, bs, true, true)
	if err == nil {
		str = GetNativeWithPrefixUrl(str)
	}
	return str, err
}

/**
 * /alioss/前缀
 *
 * @param fileName
 * @param file
 */
func (c *ossClient) UploadPrivateFileWithNativeFullUrl(fileName string, file *os.File) (string, error) {
	str, err := c.UploadFileFile(fileName, file, true, true)
	if err == nil {
		str = GetNativeWithPrefixUrl(str)
	}
	return str, err
}

func (c *ossClient) UploadPublicFileWithFullUrl(fileName string, file *os.File) (string, error) {
	str, err := c.UploadFileFile(fileName, file, false, true)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

func (c *ossClient) UploadPrivateFileWithFullUrl(fileName string, file *os.File) (string, error) {
	str, err := c.UploadFileFile(fileName, file, true, true)
	if err == nil {
		str = c.GetFullUrl(str)
	}
	return str, err
}

func (c *ossClient) UploadPrivateFile(fileName string, file *os.File) (string, error) {
	return c.UploadFileFile(fileName, file, true, true)
}

func (c *ossClient) UploadPublicFileInputStream(fileName string, file *os.File) (string, error) {
	return c.UploadFileFile(fileName, file, false, true)
}
