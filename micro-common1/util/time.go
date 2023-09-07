package util

import (
	"database/sql/driver"
	"time"
)

type Time time.Time

const (
	timeFormart = "2006-01-02 15:04:05"

	DateFormart = "2006-01-02"
)

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 0 || len(data) == 2 {
		*t = Time{}
		return
	}
	format := timeFormart
	if len(data) == len(DateFormart)+2 {
		format = DateFormart
	}
	now, err := time.ParseInLocation(`"`+format+`"`, string(data), time.Local)
	*t = Time(now)
	return
}

func (t Time) MarshalJSON() ([]byte, error) {

	if t == (Time{}) {
		return []byte{'"', '"'}, nil
	}
	b := make([]byte, 0, len(timeFormart)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormart)
	b = append(b, '"')

	if string(b) == "\"0001-01-01 00:00:00\"" {
		return []byte{'"', '"'}, nil
	}
	return b, nil
}

func (t Time) UnixNano() int64 {
	return time.Time(t).UnixNano()
}

func (t Time) Format(s string) string {
	return time.Time(t).Format(s)
}

func (t Time) Value() (driver.Value, error) {

	tTime := time.Time(t)
	return tTime.Format("2006-01-02 15:04:05"), nil
}

func (t Time) After(s Time) bool {
	return time.Time(t).After(time.Time(s))
}

func Now() Time {
	return Time(time.Now())
}

type Config struct {
	T Time
}
