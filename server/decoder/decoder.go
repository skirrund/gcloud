package decoder

import (
	"strings"

	"github.com/skirrund/gcloud/logger"

	"github.com/json-iterator/go"
)

const (
	MEDIA_JSON  = "application/json"
	MEDIA_PLAIN = "text/plain"
	MEDIA_HTML  = "text/html"
	MEDIA_XML   = "text/xml"
)

type Decoder interface {
	DecoderObj(resp []byte, obj interface{}) (Decoder, error)
}
type StringDecoder struct{}
type StreamDecoder struct{}
type JSONDecoder struct{}

var jsonDecoder = JSONDecoder{}
var stringDecoder = StringDecoder{}
var streamDecoder = StreamDecoder{}

func (d StreamDecoder) DecoderObj(resp []byte, obj interface{}) (Decoder, error) {
	if bs, ok := obj.(*[]byte); ok {
		*bs = make([]byte, len(resp))
		copy(*bs, resp)
		return d, nil
	}
	return d, nil
}

func (d StringDecoder) DecoderObj(resp []byte, obj interface{}) (Decoder, error) {
	if str, ok := obj.(*string); ok {
		*str = string(resp)
		return d, nil
	} else if bs, ok := obj.(*[]byte); ok {
		*bs = make([]byte, len(resp))
		copy(*bs, resp)
		return d, nil
	} else if _, ok := obj.([]byte); ok {
		return d, nil
	} else {
		logger.Info("[decoder else]")
		return d, nil
	}

}

func (d JSONDecoder) DecoderObj(resp []byte, obj interface{}) (Decoder, error) {
	if str, ok := obj.(*string); ok {
		*str = string(resp)
		return d, nil
	} else if bs, ok := obj.(*[]byte); ok {
		*bs = make([]byte, len(resp))
		copy(*bs, resp)
		return d, nil
	}
	err := jsoniter.Unmarshal(resp, obj)
	if err != nil {
		logger.Info("[http] JSONDecoder error:", err.Error())
	}
	return d, err
}

func GetDecoder(ct string) Decoder {
	ct = strings.ToLower(ct)
	if strings.Contains(ct, MEDIA_JSON) {
		return jsonDecoder
	} else if strings.Contains(ct, MEDIA_PLAIN) {
		return stringDecoder
	} else if strings.Contains(ct, MEDIA_HTML) {
		return stringDecoder
	} else if strings.Contains(ct, MEDIA_XML) {
		return stringDecoder
	} else {
		return streamDecoder
	}
}
