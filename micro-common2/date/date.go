package date

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"
)

type Date time.Time

const NormalDateFormat = "2006-01-02"

func (t Date) MarshalJSON() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	if t.IsZero() {
		b := []byte(`""`)
		return b, nil
	}
	b := make([]byte, 0, len(NormalDateFormat)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, NormalDateFormat)
	b = append(b, '"')
	return b, nil
}

func (t *Date) UnmarshalJSON(value []byte) error {
	var v = strings.TrimSpace(strings.Trim(string(value), "\""))
	if v == "" {
		return nil
	}
	tm, err := time.ParseInLocation(NormalDateFormat, v, time.Local)
	if err != nil {
		return err
	}
	*t = Date(Date(tm).OfTime(0, 0, 0).OfNanosecond(0))
	return nil
}

func (t Date) MarshalText() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalText: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(NormalDateFormat))
	return t.AppendFormat(b, NormalDateFormat), nil
}

func (t *Date) UnmarshalText(data []byte) error {
	*t = t.FromString(string(data))
	return nil
}

func (t Date) FromString(str string) Date {
	return ParseDate(str)
}

func ParseDate(str string) Date {
	str = dateStrFormat(str)
	return ParseDateFormat(str, NormalDateFormat)
}
func ParseDateFormat(str string, format string) Date {
	str = strings.TrimSpace(str)
	tm, err := time.ParseInLocation(format, str, time.Local)
	if nil != err {
		return Unix(0, 0).ToDate()
	}
	return Date(Date(tm).OfTime(0, 0, 0).OfNanosecond(0))
}

func (t Date) String() string {
	return t.Format(NormalDateFormat)
}

func (t Date) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}

	return t.Format(NormalDateFormat), nil
}

func (t *Date) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	*t = Date(value.(time.Time))
	return nil
}

func (t Date) Format(layout string) string {
	layout = strings.TrimSpace(layout)
	return t.T().Format(layout)
}

func (t Date) AppendFormat(b []byte, layout string) []byte {
	return t.T().AppendFormat(b, layout)
}

// 当前时间
func Today() Date {
	return Date(Date(time.Now()).OfTime(0, 0, 0).OfNanosecond(0))
}

// 构造date.Time
func NewDate(year int, month time.Month, day int, loc *time.Location) Date {
	return Date(time.Date(year, month, day, 0, 0, 0, 0, loc))
}

// 转date.Datetime
func (t Date) ToDatetime() Datetime {
	return Datetime(t)
}

// 转date.Date
func AsDate(tm time.Time) Date {
	return Date(Date(tm).OfTime(0, 0, 0).OfNanosecond(0))
}

// 转date.Time
func (t Date) ToTime() Time {
	return Time(t)
}

// 转time.Time
func (t Date) T() time.Time {
	return time.Time(t)
}

// 是否晚于
func (t Date) After(u Date) bool {
	return t.T().After(u.T())
}

// 是否早于
func (t Date) Before(u Date) bool {
	return t.T().Before(u.T())
}

// 是否等于
func (t Date) Equal(u Date) bool {
	return t.T().Equal(u.T())
}

// 是否零值
func (t Date) IsZero() bool {
	return t.T().IsZero()
}

// 获取年、月、日
func (t Date) Date() (year int, month time.Month, day int) {
	return t.T().Date()
}

// 获取年
func (t Date) Year() int {
	return t.T().Year()
}

// 获取月
func (t Date) Month() time.Month {
	return t.T().Month()
}

// 获取日
func (t Date) Day() int {
	return t.T().Day()
}

// 获取星期几
func (t Date) Weekday() time.Weekday {
	return t.T().Weekday()
}

// 获取年、第几周
func (t Date) ISOWeek() (year, week int) {
	return t.T().ISOWeek()
}

// 获取时、分、秒
func (t Date) Clock() (hour, min, sec int) {
	return t.T().Clock()
}

// 获取小时
func (t Date) Hour() int {
	return t.T().Hour()
}

// 获取分钟
func (t Date) Minute() int {
	return t.T().Minute()
}

// 获取秒
func (t Date) Second() int {
	return t.T().Second()
}

// 获取毫秒
func (t Date) Millisecond() int {
	return int(time.Duration(t.T().Nanosecond()) / time.Millisecond)
}

// 获取微秒
func (t Date) Microsecond() int {
	return int(time.Duration(t.T().Nanosecond()) / time.Microsecond)
}

// 获取纳秒
func (t Date) Nanosecond() int {
	return t.T().Nanosecond()
}

// 获取是一年中第几天
func (t Date) YearDay() int {
	return t.T().YearDay()
}

// 获取该时间 - 参数时间 的时间差
func (t Date) Sub(u Date) time.Duration {
	return t.T().Sub(u.T())
}

// 加一个时间差
func (t Date) Add(d time.Duration) Datetime {
	return Datetime(t.T().Add(d))
}

// 加年、月、日
func (t Date) AddDate(years int, months int, days int) Date {
	return Date(t.T().AddDate(years, months, days))
}

// 加时、分、秒
func (t Date) AddTime(hours int, minutes int, seconds int) Datetime {
	d := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	return t.Add(d)
}

// 加年
func (t Date) AddYears(years int) Date {
	return t.AddDate(years, 0, 0)
}

// 加月
func (t Date) AddMonths(months int) Date {
	return t.AddDate(0, months, 0)
}

// 加日
func (t Date) AddDays(days int) Date {
	return t.AddDate(0, 0, days)
}

// 加小时
func (t Date) AddHours(hours int) Datetime {
	d := time.Duration(hours) * time.Hour
	return t.Add(d)
}

// 加分钟
func (t Date) AddMinutes(minutes int) Datetime {
	d := time.Duration(minutes) * time.Minute
	return t.Add(d)
}

// 加秒
func (t Date) AddSeconds(seconds int) Datetime {
	d := time.Duration(seconds) * time.Second
	return t.Add(d)
}

// 加纳秒
func (t Date) AddNanoseconds(nanoseconds int) Datetime {
	d := time.Duration(nanoseconds) * time.Nanosecond
	return t.Add(d)
}

// 指定时间
func (t Date) Of(year int, month time.Month, day int, hour int, minute int, second int, nanosecond int) Datetime {
	return Datetime(time.Date(year, month, day, hour, minute, second, nanosecond, t.Location()))
}

// 指定年、月、日
func (t Date) OfDate(year int, month time.Month, day int) Date {
	hour, minute, second := t.Clock()
	return Date(t.Of(year, month, day, hour, minute, second, t.Nanosecond()))
}

// 指定时、分、秒
func (t Date) OfTime(hour int, minute int, second int) Datetime {
	year, month, day := t.Date()
	return t.Of(year, month, day, hour, minute, second, t.Nanosecond())
}

// 指定年
func (t Date) OfYear(year int) Date {
	return t.OfDate(year, t.Month(), t.Day())
}

// 指定月
func (t Date) OfMonth(month time.Month) Date {
	return t.OfDate(t.Year(), month, t.Day())
}

// 指定日
func (t Date) OfDay(day int) Date {
	return t.OfDate(t.Year(), t.Month(), day)
}

// 指定小时
func (t Date) OfHour(hour int) Datetime {
	return t.OfTime(hour, t.Minute(), t.Second())
}

// 指定分钟
func (t Date) OfMinute(minute int) Datetime {
	return t.OfTime(t.Hour(), minute, t.Second())
}

// 指定秒
func (t Date) OfSecond(second int) Datetime {
	return t.OfTime(t.Hour(), t.Minute(), second)
}

// 指定纳秒
func (t Date) OfNanosecond(nanosecond int) Datetime {
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	return t.Of(year, month, day, hour, minute, second, nanosecond)
}

// 转UTC时间
func (t Date) UTC() Datetime {
	return Datetime(t.T().UTC())
}

// 转本地时间
func (t Date) Local() Datetime {
	return Datetime(t.T().Local())
}

// 转指定时区时间
func (t Date) In(loc *time.Location) Datetime {
	return Datetime(t.T().In(loc))
}

// 获取时区
func (t Date) Location() *time.Location {
	return t.T().Location()
}

// 获取时区
func (t Date) Zone() (name string, offset int) {
	return t.T().Zone()
}

// 获取UTC时间戳
func (t Date) Unix() int64 {
	return t.T().Unix()
}

// 获取UTC纳秒数
func (t Date) UnixNano() int64 {
	return t.T().Unix()
}

// 从目标日期开始到现在的时间差
func SinceDate(t Date) time.Duration {
	return Today().Sub(t)
}

// 现在到目标日期的时间差
func UntilDate(t Date) time.Duration {
	return t.Sub(Today())
}
