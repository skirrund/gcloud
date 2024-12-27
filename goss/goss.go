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
	GetNativePrefix() string
	GetFullUrlWithSign(fileName string, expiredInSec int64) (string, error)
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

type CommonOss struct {
	ossClient OssClient
}

var defaultOss *CommonOss

var ContentTypes = make(map[string]string)

func init() {
	ContentTypes[".bmp"] = "image/bmp"
	ContentTypes[".tif"] = "image/tiff"
	ContentTypes[".tiff"] = "image/tiff"
	ContentTypes[".gif"] = "image/gif"
	ContentTypes[".jpeg"] = "image/jpeg"
	ContentTypes[".png"] = "image/png"
	ContentTypes[".jpg"] = "image/jpeg"
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

func GetcontentType(fileName string) string {
	index := strings.Index(fileName, ".")
	if index == -1 {
		return "image/jpeg"
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

func SubStringBlackSlash(s string) string {
	if !strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "\\") {
		return s
	} else {
		s = s[1:]
	}
	return SubStringBlackSlash(s)
}

func (co CommonOss) GetNativeWithPrefixUrl(fileName string) string {
	return co.ossClient.GetNativePrefix() + fileName
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

func GetDefault(c OssClient) (*CommonOss, error) {
	if defaultOss != nil {
		return defaultOss, nil
	}
	return NewDefault(c)
}

func NewDefault(c OssClient) (*CommonOss, error) {
	cli, err := c.NewDefaultClient()
	if err != nil {
		return nil, err
	}
	defaultOss = &CommonOss{
		ossClient: cli,
	}
	return defaultOss, nil
}
