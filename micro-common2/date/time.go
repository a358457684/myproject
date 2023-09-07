package date

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"
)

type Time time.Time

const NormalTimeFormat = "15:04:05"

func (t Time) MarshalJSON() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	if t.IsZero() {
		b := []byte(`""`)
		return b, nil
	}
	b := make([]byte, 0, len(NormalTimeFormat)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, NormalTimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t *Time) UnmarshalJSON(value []byte) error {
	var v = strings.TrimSpace(strings.Trim(string(value), "\""))
	if v == "" {
		return nil
	}
	tm, err := time.ParseInLocation(NormalTimeFormat, v, time.Local)
	if err != nil {
		return err
	}
	*t = Time(Time(tm).OfDate(1, time.January, 1))
	return nil
}

func (t Time) MarshalText() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalText: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(NormalTimeFormat))
	return t.AppendFormat(b, NormalTimeFormat), nil
}

func (t *Time) UnmarshalText(data []byte) error {
	*t = t.FromString(string(data))
	return nil
}

func (t Time) FromString(str string) Time {
	return ParseTime(str)
}

func ParseTime(str string) Time {
	str = dateStrFormat(str)
	return ParseTimeFormat(str, NormalTimeFormat)
}
func ParseTimeFormat(str string, format string) Time {
	str = strings.TrimSpace(str)

	tm, err := time.ParseInLocation(format, str, time.Local)

	if nil != err {
		return Unix(0, 0).ToTime()
	}
	return Time(Time(tm).OfDate(1, time.January, 1))
}

func (t Time) String() string {
	return t.Format(NormalTimeFormat)
}

func (t Time) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}

	return t.Format(NormalTimeFormat), nil
}

func (t *Time) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	*t = Time(value.(time.Time))
	return nil
}

func (t Time) Format(layout string) string {
	layout = strings.TrimSpace(layout)
	return t.T().Format(layout)
}

func (t Time) AppendFormat(b []byte, layout string) []byte {
	return t.T().AppendFormat(b, layout)
}

// 当前时间
func CurrentTime() Time {
	return Time(Time(time.Now()).OfDate(1, time.January, 1))
}

// 构造date.Time
func NewTime(hour, min, sec, nsec int, loc *time.Location) Time {
	return Time(time.Date(1, time.January, 1, hour, min, sec, nsec, loc))
}

// 转date.Datetime
func (t Time) ToDatetime() Datetime {
	return Datetime(t)
}

// 转date.Date
func (t Time) ToDate() Date {
	return Date(t)
}

// 转date.Time
func AsTime(tm time.Time) Time {
	return Time(Time(tm).OfDate(1, time.January, 1))
}

// 转time.Time
func (t Time) T() time.Time {
	return time.Time(t)
}

// 是否晚于
func (t Time) After(u Time) bool {
	return t.T().After(u.T())
}

// 是否早于
func (t Time) Before(u Time) bool {
	return t.T().Before(u.T())
}

// 是否等于
func (t Time) Equal(u Time) bool {
	return t.T().Equal(u.T())
}

// 是否零值
func (t Time) IsZero() bool {
	return t.T().IsZero()
}

// 获取年、月、日
func (t Time) Date() (year int, month time.Month, day int) {
	return t.T().Date()
}

// 获取年
func (t Time) Year() int {
	return t.T().Year()
}

// 获取月
func (t Time) Month() time.Month {
	return t.T().Month()
}

// 获取日
func (t Time) Day() int {
	return t.T().Day()
}

// 获取星期几
func (t Time) Weekday() time.Weekday {
	return t.T().Weekday()
}

// 获取年、第几周
func (t Time) ISOWeek() (year, week int) {
	return t.T().ISOWeek()
}

// 获取时、分、秒
func (t Time) Clock() (hour, min, sec int) {
	return t.T().Clock()
}

// 获取小时
func (t Time) Hour() int {
	return t.T().Hour()
}

// 获取分钟
func (t Time) Minute() int {
	return t.T().Minute()
}

// 获取秒
func (t Time) Second() int {
	return t.T().Second()
}

// 获取毫秒
func (t Time) Millisecond() int {
	return int(time.Duration(t.T().Nanosecond()) / time.Millisecond)
}

// 获取微秒
func (t Time) Microsecond() int {
	return int(time.Duration(t.T().Nanosecond()) / time.Microsecond)
}

// 获取纳秒
func (t Time) Nanosecond() int {
	return t.T().Nanosecond()
}

// 获取是一年中第几天
func (t Time) YearDay() int {
	return t.T().YearDay()
}

// 获取该时间 - 参数时间 的时间差
func (t Time) Sub(u Time) time.Duration {
	return t.T().Sub(u.T())
}

// 加一个时间差
func (t Time) Add(d time.Duration) Datetime {
	return Datetime(t.T().Add(d))
}

// 加年、月、日
func (t Time) AddDate(years int, months int, days int) Datetime {
	return Datetime(t.T().AddDate(years, months, days))
}

// 加时、分、秒
func (t Time) AddTime(hours int, minutes int, seconds int) Datetime {
	d := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	return t.Add(d)
}

// 加年
func (t Time) AddYears(years int) Datetime {
	return t.AddDate(years, 0, 0)
}

// 加月
func (t Time) AddMonths(months int) Datetime {
	return t.AddDate(0, months, 0)
}

// 加日
func (t Time) AddDays(days int) Datetime {
	return t.AddDate(0, 0, days)
}

// 加小时
func (t Time) AddHours(hours int) Datetime {
	d := time.Duration(hours) * time.Hour
	return t.Add(d)
}

// 加分钟
func (t Time) AddMinutes(minutes int) Datetime {
	d := time.Duration(minutes) * time.Minute
	return t.Add(d)
}

// 加秒
func (t Time) AddSeconds(seconds int) Datetime {
	d := time.Duration(seconds) * time.Second
	return t.Add(d)
}

// 加纳秒
func (t Time) AddNanoseconds(nanoseconds int) Datetime {
	d := time.Duration(nanoseconds) * time.Nanosecond
	return t.Add(d)
}

// 指定时间
func (t Time) Of(year int, month time.Month, day int, hour int, minute int, second int, nanosecond int) Datetime {
	return Datetime(time.Date(year, month, day, hour, minute, second, nanosecond, t.Location()))
}

// 指定年、月、日
func (t Time) OfDate(year int, month time.Month, day int) Datetime {
	hour, minute, second := t.Clock()
	return t.Of(year, month, day, hour, minute, second, t.Nanosecond())
}

// 指定时、分、秒
func (t Time) OfTime(hour int, minute int, second int) Time {
	year, month, day := t.Date()
	return Time(t.Of(year, month, day, hour, minute, second, t.Nanosecond()))
}

// 指定年
func (t Time) OfYear(year int) Datetime {
	return t.OfDate(year, t.Month(), t.Day())
}

// 指定月
func (t Time) OfMonth(month time.Month) Datetime {
	return t.OfDate(t.Year(), month, t.Day())
}

// 指定日
func (t Time) OfDay(day int) Datetime {
	return t.OfDate(t.Year(), t.Month(), day)
}

// 指定小时
func (t Time) OfHour(hour int) Time {
	return t.OfTime(hour, t.Minute(), t.Second())
}

// 指定分钟
func (t Time) OfMinute(minute int) Time {
	return t.OfTime(t.Hour(), minute, t.Second())
}

// 指定秒
func (t Time) OfSecond(second int) Time {
	return t.OfTime(t.Hour(), t.Minute(), second)
}

// 指定纳秒
func (t Time) OfNanosecond(nanosecond int) Time {
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	return Time(t.Of(year, month, day, hour, minute, second, nanosecond))
}

// 转UTC时间
func (t Time) UTC() Datetime {
	return Datetime(t.T().UTC())
}

// 转本地时间
func (t Time) Local() Datetime {
	return Datetime(t.T().Local())
}

// 转指定时区时间
func (t Time) In(loc *time.Location) Datetime {
	return Datetime(t.T().In(loc))
}

// 获取时区
func (t Time) Location() *time.Location {
	return t.T().Location()
}

// 获取时区
func (t Time) Zone() (name string, offset int) {
	return t.T().Zone()
}

// 获取UTC时间戳
func (t Time) Unix() int64 {
	return t.T().Unix()
}

// 获取UTC纳秒数
func (t Time) UnixNano() int64 {
	return t.T().Unix()
}

// 从目标时间开始到现在的时间差
func SinceTime(t Time) time.Duration {
	return CurrentTime().Sub(t)
}

// 现在到目标时间的时间差
func UntilTime(t Time) time.Duration {
	return t.Sub(CurrentTime())
}
