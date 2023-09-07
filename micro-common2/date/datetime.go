package date

import (
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"time"
)

type Datetime time.Time

const NormalDatetimeFormat = "2006-01-02 15:04:05"

var (
	datePattern *regexp.Regexp
)

func init() {
	datePattern, _ = regexp.Compile(`([-:\s])(\d)([-:\s])`)
}

func (t Datetime) MarshalJSON() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	if t.IsZero() {
		b := []byte(`""`)
		return b, nil
	}
	b := make([]byte, 0, len(NormalDatetimeFormat)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, NormalDatetimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t *Datetime) UnmarshalJSON(value []byte) error {
	var v = strings.TrimSpace(strings.Trim(string(value), "\""))
	if v == "" {
		return nil
	}
	tm, err := time.ParseInLocation(NormalDatetimeFormat, v, time.Local)
	if err != nil {
		return err
	}
	*t = Datetime(tm)
	return nil
}

func (t Datetime) MarshalText() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalText: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(NormalDatetimeFormat))
	return t.AppendFormat(b, NormalDatetimeFormat), nil
}

func (t *Datetime) UnmarshalText(data []byte) error {
	*t = t.FromString(string(data))
	return nil
}

func (t Datetime) FromString(str string) Datetime {
	return ParseDatetime(str)
}

func ParseDatetime(str string) Datetime {
	str = dateStrFormat(str)
	return ParseDatetimeFormat(str, NormalDatetimeFormat)
}
func ParseDatetimeFormat(str, format string) Datetime {
	str = strings.TrimSpace(str)
	tm, err := time.ParseInLocation(format, str, time.Local)
	if nil != err {
		return Unix(0, 0)
	}
	return Datetime(tm)
}

func (t Datetime) String() string {
	return t.Format(NormalDatetimeFormat)
}

func (t Datetime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}

	return t.Format(NormalDatetimeFormat), nil
}

func (t *Datetime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	*t = Datetime(value.(time.Time))
	return nil
}

func (t Datetime) Format(layout string) string {
	layout = strings.TrimSpace(layout)
	return t.T().Format(layout)
}

func (t Datetime) AppendFormat(b []byte, layout string) []byte {
	return t.T().AppendFormat(b, layout)
}

// 当前时间
func Now() Datetime {
	return Datetime(time.Now())
}

// 构造date.Datetime
func NewDatetime(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) Datetime {
	return Datetime(time.Date(year, month, day, hour, min, sec, nsec, loc))
}

// 转date.Time
func AsDatetime(tm time.Time) Datetime {
	return Datetime(tm)
}

// 转date.Date
func (t Datetime) ToDate() Date {
	return Date(t.OfTime(0, 0, 0).OfNanosecond(0))
}

// 转date.Time
func (t Datetime) ToTime() Time {
	return Time(t.OfDate(1, time.January, 1))
}

// 转time.Time
func (t Datetime) T() time.Time {
	return time.Time(t)
}

// 是否晚于
func (t Datetime) After(u Datetime) bool {
	return t.T().After(u.T())
}

// 是否早于
func (t Datetime) Before(u Datetime) bool {
	return t.T().Before(u.T())
}

// 是否等于
func (t Datetime) Equal(u Datetime) bool {
	return t.T().Equal(u.T())
}

// 是否零值
func (t Datetime) IsZero() bool {
	return t.T().IsZero()
}

// 获取年、月、日
func (t Datetime) Date() (year int, month time.Month, day int) {
	return t.T().Date()
}

// 获取年
func (t Datetime) Year() int {
	return t.T().Year()
}

// 获取月
func (t Datetime) Month() time.Month {
	return t.T().Month()
}

// 获取日
func (t Datetime) Day() int {
	return t.T().Day()
}

// 获取星期几
func (t Datetime) Weekday() time.Weekday {
	return t.T().Weekday()
}

// 获取年、第几周
func (t Datetime) ISOWeek() (year, week int) {
	return t.T().ISOWeek()
}

// 获取时、分、秒
func (t Datetime) Clock() (hour, min, sec int) {
	return t.T().Clock()
}

// 获取小时
func (t Datetime) Hour() int {
	return t.T().Hour()
}

// 获取分钟
func (t Datetime) Minute() int {
	return t.T().Minute()
}

// 获取秒
func (t Datetime) Second() int {
	return t.T().Second()
}

// 获取毫秒
func (t Datetime) Millisecond() int {
	return int(time.Duration(t.T().Nanosecond()) / time.Millisecond)
}

// 获取微秒
func (t Datetime) Microsecond() int {
	return int(time.Duration(t.T().Nanosecond()) / time.Microsecond)
}

// 获取纳秒
func (t Datetime) Nanosecond() int {
	return t.T().Nanosecond()
}

// 获取是一年中第几天
func (t Datetime) YearDay() int {
	return t.T().YearDay()
}

// 获取该时间 - 参数时间 的时间差
func (t Datetime) Sub(u Datetime) time.Duration {
	return t.T().Sub(u.T())
}

// 加一个时间差
func (t Datetime) Add(d time.Duration) Datetime {
	return Datetime(t.T().Add(d))
}

// 加年、月、日
func (t Datetime) AddDate(years int, months int, days int) Datetime {
	return Datetime(t.T().AddDate(years, months, days))
}

// 加时、分、秒
func (t Datetime) AddTime(hours int, minutes int, seconds int) Datetime {
	d := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	return t.Add(d)
}

// 加年
func (t Datetime) AddYears(years int) Datetime {
	return t.AddDate(years, 0, 0)
}

// 加月
func (t Datetime) AddMonths(months int) Datetime {
	return t.AddDate(0, months, 0)
}

// 加日
func (t Datetime) AddDays(days int) Datetime {
	return t.AddDate(0, 0, days)
}

// 加小时
func (t Datetime) AddHours(hours int) Datetime {
	d := time.Duration(hours) * time.Hour
	return t.Add(d)
}

// 加分钟
func (t Datetime) AddMinutes(minutes int) Datetime {
	d := time.Duration(minutes) * time.Minute
	return t.Add(d)
}

// 加秒
func (t Datetime) AddSeconds(seconds int) Datetime {
	d := time.Duration(seconds) * time.Second
	return t.Add(d)
}

// 加纳秒
func (t Datetime) AddNanoseconds(nanoseconds int) Datetime {
	d := time.Duration(nanoseconds) * time.Nanosecond
	return t.Add(d)
}

// 指定时间
func (t Datetime) Of(year int, month time.Month, day int, hour int, minute int, second int, nanosecond int) Datetime {
	return Datetime(time.Date(year, month, day, hour, minute, second, nanosecond, t.Location()))
}

// 指定年、月、日
func (t Datetime) OfDate(year int, month time.Month, day int) Datetime {
	hour, minute, second := t.Clock()
	return t.Of(year, month, day, hour, minute, second, t.Nanosecond())
}

// 指定时、分、秒
func (t Datetime) OfTime(hour int, minute int, second int) Datetime {
	year, month, day := t.Date()
	return t.Of(year, month, day, hour, minute, second, t.Nanosecond())
}

// 指定年
func (t Datetime) OfYear(year int) Datetime {
	return t.OfDate(year, t.Month(), t.Day())
}

// 指定月
func (t Datetime) OfMonth(month time.Month) Datetime {
	return t.OfDate(t.Year(), month, t.Day())
}

// 指定日
func (t Datetime) OfDay(day int) Datetime {
	return t.OfDate(t.Year(), t.Month(), day)
}

// 指定小时
func (t Datetime) OfHour(hour int) Datetime {
	return t.OfTime(hour, t.Minute(), t.Second())
}

// 指定分钟
func (t Datetime) OfMinute(minute int) Datetime {
	return t.OfTime(t.Hour(), minute, t.Second())
}

// 指定秒
func (t Datetime) OfSecond(second int) Datetime {
	return t.OfTime(t.Hour(), t.Minute(), second)
}

// 指定纳秒
func (t Datetime) OfNanosecond(nanosecond int) Datetime {
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	return t.Of(year, month, day, hour, minute, second, nanosecond)
}

// 转UTC时间
func (t Datetime) UTC() Datetime {
	return Datetime(t.T().UTC())
}

// 转本地时间
func (t Datetime) Local() Datetime {
	return Datetime(t.T().Local())
}

// 转指定时区时间
func (t Datetime) In(loc *time.Location) Datetime {
	return Datetime(t.T().In(loc))
}

// 获取时区
func (t Datetime) Location() *time.Location {
	return t.T().Location()
}

// 获取时区
func (t Datetime) Zone() (name string, offset int) {
	return t.T().Zone()
}

// 获取UTC时间戳
func (t Datetime) Unix() int64 {
	return t.T().Unix()
}

// 获取UTC纳秒数
func (t Datetime) UnixNano() int64 {
	return t.T().Unix()
}

// 从目标时间开始到现在的时间差
func Since(t Datetime) time.Duration {
	return Now().Sub(t)
}

// 现在到目标时间的时间差
func Until(t Datetime) time.Duration {
	return t.Sub(Now())
}

// UTC时间戳和纳秒数转时间
func Unix(sec int64, nsec int64) Datetime {
	return Datetime(time.Unix(sec, nsec))
}

func dateStrFormat(input string) string {
	input = strings.Replace(strings.TrimSpace(input), "/", "-", -1)
	if strings.Index(input, ":") > 0 && strings.Index(input, ":") == strings.LastIndex(input, ":") {
		input = input + ":00"
	}
	for {
		if datePattern.MatchString(input) {
			input = datePattern.ReplaceAllString(input, `${1}0${2}${3}`)
		} else {
			break
		}
	}
	if strings.Index(input, ":") == 1 {
		input = "0" + input
	}
	length := len(input)
	if length > 2 {
		if strings.LastIndex(input, ":") == length-2 {
			input = input[0:length-1] + "0" + input[length-1:]
		} else if strings.LastIndex(input, "-") == length-2 {
			input = input[0:length-1] + "0" + input[length-1:]
		}
	}
	return input
}
