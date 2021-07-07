package file

import (
	"encoding/base64"
	"io/fs"
	"os"
	"path"
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
