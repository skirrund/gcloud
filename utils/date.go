package utils

import (
	"fmt"
	"time"

	"github.com/skirrund/gcloud/logger"

	"database/sql/driver"
)

type DateTime struct {
	Time time.Time
}

const (
	TimeFormat     = "2006-01-02 15:04:05"
	TimeFormatZero = "0001-01-01 00:00:00"
)

func (t DateTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t.Time).Format(TimeFormat))
	return []byte(stamp), nil
}

func (t DateTime) IsZero() bool {
	return t.Time.IsZero()
}

func (t *DateTime) UnmarshalJSON(b []byte) error {
	str := string(b)
	pt, err := time.ParseInLocation(`"`+TimeFormat+`"`, str, time.Local)
	if err != nil {
		logger.Error("[DateTime] format error:" + str)
		return err
	}
	*t = DateTime{
		Time: pt,
	}
	return nil
}

func (t DateTime) Format() string {
	return fmt.Sprintf("\"%s\"", time.Time(t.Time).Format(TimeFormat))
}

func (t *DateTime) Scan(value interface{}) error {
	if v, ok := value.(time.Time); ok {
		dt := DateTime{
			Time: v,
		}
		*t = dt
	}
	return nil
}

func (t DateTime) String() string {
	return time.Time(t.Time).Format(TimeFormat)
}

func (t DateTime) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return time.Time(t.Time), nil
}
