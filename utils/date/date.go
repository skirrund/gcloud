package date

import (
	"time"
)

const (
	DefaultTimeFormat = "2006-01-02 15:04:05"
	TimeFormatDate    = "2006-01-02"
)

//获取明天零点的时间
func GetBetweenNextDaySeconds() time.Duration {
	now := time.Now()
	nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return time.Duration(nextDay.Unix()-now.Unix()) * time.Second
}

func Milliseconds(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

func ToString(t time.Time, formater string) string {
	return t.Format(formater)
}
