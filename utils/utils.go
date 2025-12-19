package utils

import (
	"bytes"
	"io"
	"math/rand"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/skirrund/gcloud/logger"

	cc "github.com/skirrund/gcloud/config"

	sonic "github.com/bytedance/sonic"
	"github.com/google/uuid"
)

var (
	localIP     string
	privateCIDR []*net.IPNet
)

const (
	REGX_ID_PATTERN_15 = "^[1-9]\\d{7}((0\\d)|(1[0-2]))(([0|1|2]\\d)|3[0-1])\\d{3}$"
	REGX_ID_PATTERN_18 = "^[1-9]\\d{5}[1-9]\\d{3}((0\\d)|(1[0-2]))(([0|1|2]\\d)|3[0-1])\\d{3}([0-9]|X)$"
)

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
	filenameExtension := SubStr(fileName, strings.LastIndex(fileName, "."), -1)
	if len(filenameExtension) > 0 {
		contentType := ContentTypes[strings.ToLower(filenameExtension)]
		if len(contentType) > 0 {
			return contentType
		}
	}
	return "application/octet-stream"
}

func IsIdNoCorrect(idNo string) bool {
	match, _ := regexp.MatchString(REGX_ID_PATTERN_15, idNo)
	if match {
		return true
	}
	match, _ = regexp.MatchString(REGX_ID_PATTERN_18, idNo)
	return match
}

func VerifyEmailFormat(email string) bool {
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// 身份证号获取年龄
func GetAgeFromIdNum(idNum string) int {
	return GetAgeFromIdNumEndTime(idNum, time.Now())
}

// 身份证号获取到某一时间点年龄
func GetAgeFromIdNumEndTime(idNum string, endTime time.Time) int {
	str := SubStr(idNum, 6, 8)
	age, err := time.ParseInLocation("20060102", str, time.Local)
	if err != nil {
		return 0
	}
	actYear := endTime.Year()
	actMonth := endTime.Month()
	actDay := endTime.Day()
	ageYear := age.Year()
	ageMonth := age.Month()
	ageDay := age.Day()
	yearInterval := actYear - ageYear
	// 如果 d1的 月-日 小于 d2的 月-日 那么 yearInterval-- 这样就得到了相差的年数
	if actMonth < ageMonth || (actMonth == ageMonth && actDay < ageDay) {
		yearInterval--
	}
	return yearInterval
}

func NewOptions(config cc.IConfig, opts any) {
	t := reflect.TypeOf(opts)
	kind := t.Kind()
	if kind == reflect.Ptr {
		t = t.Elem()
	}
	if kind == reflect.Struct {
		logger.Error("[common] NewOptions: check type error not Struct")
		return
	}
	bean := reflect.ValueOf(opts)
	if bean.IsZero() {
		return
	}
	bean = bean.Elem()

	num := t.NumField()
	var f reflect.StructField
	for i := 0; i != num; i++ {
		f = t.Field(i)
		v := bean.Field(i)
		tag := f.Tag.Get("property")
		if v.IsValid() && v.CanSet() {
			switch v.Kind() {
			case reflect.String:
				v.SetString(config.GetString(tag))
			case reflect.Slice:
				s := reflect.ValueOf(config.GetStringSlice(tag))
				v.Set(s)
			case reflect.Int:
				v.SetInt(config.GetInt64(tag))
			case reflect.Int8:
				v.SetInt(config.GetInt64(tag))
			case reflect.Int16:
				v.SetInt(config.GetInt64(tag))
			case reflect.Int32:
				v.SetInt(config.GetInt64(tag))
			case reflect.Int64:
				v.SetInt(config.GetInt64(tag))
			case reflect.Bool:
				v.SetBool(config.GetBool(tag))
			case reflect.Float32:
				v.SetFloat(config.GetFloat64(tag))
			}
		}
	}
	logger.Info("[common]  NewOptions: create opts:", opts)
}

func Mask(str string, before int, after int) string {
	str = strings.TrimSpace(str)
	chs := []rune(str)
	l := len(chs)
	if l == 0 {
		return str
	} else if l <= before+after {
		return str
	} else {
		i := before
		for k := l - after; i < k; i++ {
			chs[i] = '*'
		}
		return string(chs)
	}
}

func Uuid() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func UnicodeIndex(str, substr string) int {
	idx := strings.Index(str, substr)
	if idx > 0 {
		prefix := []byte(str)[0:idx]
		rs := []rune(string(prefix))
		idx = len(rs)
	}
	return idx
}

func UnicodeLastIndex(str, substr string) int {
	idx := strings.LastIndex(str, substr)
	if idx > 0 {
		prefix := []byte(str)[0:idx]
		rs := []rune(string(prefix))
		idx = len(rs)
	}
	return idx
}

// 截取字符串，支持多字节字符
// start：起始下标，负数从从尾部开始，最后一个为-1
// length：截取长度，负数表示截取到末尾
func SubStr(str string, start int, length int) (result string) {
	s := []rune(str)
	total := len(s)
	if total == 0 {
		return
	}
	// 允许从尾部开始计算
	if start < 0 {
		start = total + start
		if start < 0 {
			return
		}
	}
	if start > total {
		return
	}
	// 到末尾
	if length < 0 {
		length = total
	}

	end := start + length
	if end > total {
		result = string(s[start:])
	} else {
		result = string(s[start:end])
	}

	return
}

// 判断obj是否在target中，target支持的类型arrary,slice,map
func Contains(obj any, target any) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}
func Contains2(obj any, target any) bool {
	if obj == nil || target == nil {
		return false
	}
	ptrObj := reflect.ValueOf(obj)
	if ptrObj.Kind() == reflect.Ptr {
		// 获取指针指向的值
		obj = ptrObj.Elem().Interface()
	}
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			var arrItem = targetValue.Index(i).Interface()
			ptrVal := reflect.ValueOf(arrItem)
			if ptrVal.Kind() == reflect.Ptr {
				// 获取指针指向的值
				arrItem = ptrVal.Elem().Interface()
			}
			if arrItem == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}

func UnmarshalFromString(str string, obj any) error {
	return sonic.UnmarshalString(str, obj)
}

func Unmarshal(bytes []byte, obj any) error {
	return sonic.Unmarshal(bytes, obj)
}

func Marshal(obj any) ([]byte, error) {
	return sonic.Marshal(obj)
}

func MarshalToString(obj any) (string, error) {
	return sonic.MarshalString(obj)
}

func GenerateCode(size int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	str := ""
	for i := 0; i < size; i++ {
		str += strconv.FormatInt(rand.Int63n(8)+1, 10)
	}
	return str
}

func GetValidDate(validTime time.Duration) time.Time {
	return time.Now().Add(validTime)
}

func ReadFile(localFile string) ([]byte, error) {
	var chunk []byte //  数据块
	file, err := os.Open(localFile)
	if err != nil {
		logger.Error(err)
		return chunk, err
	}
	defer file.Close()
	for {

		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			logger.Error(err)
			return chunk, err
		}
		if n == 0 {
			break
		}

		chunk = append(chunk, buffer[:n]...)
	}
	return chunk, err
}

func GetStringParamsMapFromUrl(paramsStr string) map[string]string {
	resultMap := make(map[string]string)
	if len(paramsStr) == 0 {
		return resultMap
	}
	index := strings.Index(paramsStr, "?")
	if index != -1 {
		paramsStr = SubStr(paramsStr, index+1, -1)
	} else {
		return resultMap
	}

	strs := strings.Split(paramsStr, "&")
	var valuePair []string
	for _, p := range strs {
		valuePair = strings.Split(p, "=")
		if len(valuePair) != 1 {
			if strings.Contains(valuePair[1], "%") {
				v, err := url.QueryUnescape(valuePair[1])
				if err == nil {
					valuePair[1] = v
				}
			}
			resultMap[valuePair[0]] = valuePair[1]
		}

	}
	return resultMap
}

func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func isFilteredIP(ip net.IP) bool {
	for _, privateIP := range privateCIDR {
		if privateIP.Contains(ip) {
			return true
		}
	}
	return false
}

func LocalIP() string {
	if localIP != "" {
		return localIP
	}

	faces, err := getFaces()
	if err != nil {
		return ""
	}

	for _, address := range faces {
		ipNet, ok := address.(*net.IPNet)
		if !ok || ipNet.IP.To4() == nil || isFilteredIP(ipNet.IP) {
			continue
		}

		localIP = ipNet.IP.String()
		break
	}

	if localIP != "" {
		logger.Infof("Local IP:%s", localIP)
	}

	return localIP
}

func getFaces() ([]net.Addr, error) {
	var upAddrs []net.Addr

	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Errorf("get Interfaces failed,err:%+v", err)
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if (iface.Flags & net.FlagLoopback) != 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			logger.Errorf("get InterfaceAddress failed,err:%+v", err)
			return nil, err
		}

		upAddrs = append(upAddrs, addresses...)
	}

	return upAddrs, nil
}
