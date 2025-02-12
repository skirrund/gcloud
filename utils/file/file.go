package file

import (
	"encoding/base64"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/skirrund/gcloud/goss"
	"github.com/skirrund/gcloud/goss/alioss"
	"github.com/skirrund/gcloud/logger"
)

func SaveImageBase64(base64Str string, filePath string, fileName string) (fn string, err error) {
	ex, err := Exist(filePath)
	if err != nil {
		return fn, err
	}
	if !ex {
		err = os.MkdirAll(filePath, fs.ModePerm)
		if err != nil {
			return fn, err
		}
	}
	f := path.Join(filePath, fileName)
	bs, err := base64.StdEncoding.DecodeString(base64Str)
	if err == nil {
		os.WriteFile(f, bs, fs.ModePerm)
	}
	return f, err
}

func Exist(localFile string) (bool, error) {
	_, err := os.Stat(localFile)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetPublicURL(filePath string, publicPath string, publicDomain string) string {
	if len(filePath) == 0 {
		logger.Warn("filePath is null , return filePath")
		return filePath
	} else {
		client, err := goss.GetDefault(alioss.OssClient{})
		if err != nil {
			return filepath.Join(publicDomain, filePath)
		}
		if strings.HasPrefix(filePath, publicPath) {
			filePath = strings.Replace(filePath, publicPath, "", 1)
		} else if strings.HasPrefix(filePath, client.GetNativePrefix()) {
			str, err := client.GetFullUrlWithSign(filePath, 600)
			if err == nil {
				return str
			}
		}
	}
	return filepath.Join(publicDomain, filePath)
}
