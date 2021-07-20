package image

import (
	"encoding/base64"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"
)

func TestCommpressBase64Pic(t *testing.T) {
	f, err := os.Open("/Users/jerry.shi/Desktop/gaitubao_WechatIMG1 2_bmp.bmp")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	t.Log("file size:", len(bs)/1024)
	if err != nil {
		t.Error(err)
	}
	str := base64.StdEncoding.EncodeToString(bs)
	bs = CommpressBase64PicToByte(str, 89, 4096)
	t.Log(">>>>>>>", len(bs)/1024)
	err = os.WriteFile("/Users/jerry.shi/Desktop/testcompress.jpeg", bs, fs.ModePerm)
	t.Error(err)
}
