package gtime

import "time"

const (
	DateTimeFormat = "2006-01-02 15:04:05"
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"

	DayDuration = time.Hour * 24
	//天秒数
	DaySec = 24 * 3600
)

var (
	OffsetDuration time.Duration
)

//当前时间
func Now() time.Time {
	t := time.Now()
	if OffsetDuration == 0 {
		return t
	}
	return t.Add(OffsetDuration)
}

func Since(t time.Time) time.Duration {
	return Now().Sub(t)
}

//格式化
func Format(t time.Time, layout string) string {
	return t.Format(layout)
}

//格式化2006-01-02 15:04:05
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

//格式化2006-01-02
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

//格式化15:04:05
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

//解析自定义时间
func Parse(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.Local)
}

//解析日期时间
func ParseDateTime(value string) (time.Time, error) {
	return time.ParseInLocation(DateTimeFormat, value, time.Local)
}

//解析日期
func ParseDate(value string) (time.Time, error) {
	return time.ParseInLocation(DateFormat, value, time.Local)
}

//解析时间
func ParseTime(value string) (time.Time, error) {
	return time.ParseInLocation(TimeFormat, value, time.Local)
}

//获取自定义时间日期
func Date(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.Local)
}

//时间戳转时间
func Unix(sec int64) time.Time {
	return time.Unix(sec, 0)
}

//转换持续时间格式
func Duration(sec int) time.Duration {
	return time.Second * time.Duration(sec)
}

//持续时间转秒数
func Second(d time.Duration) int {
	return int(d / time.Second)
}

//获取指定某天的0点0分0秒时间
func GetZero(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

//获取当天过了多少秒
func GetDaySec(t time.Time) int {
	return t.Hour()*60*60 + t.Minute()*60 + t.Second()
}

//是否相同天
func IsSameDay(t1, t2 time.Time) bool {
	t1, t2 = GetZero(t1), GetZero(t2)
	return t1.Unix() == t2.Unix()
}

//是否相同小时
func IsSameHour(t1, t2 time.Time) bool {
	return t1.Hour() == t2.Hour() && t1.Day() == t2.Day() && t1.Month() == t2.Month() && t1.Year() == t2.Year()
}

//t1-t2
func GetDeltaDays(t1, t2 time.Time) int {
	return int(GetZero(t1).Sub(GetZero(t2)))
}

//获取周1～周7
func GetWeekDay(t time.Time) int {
	t = GetZero(t)
	week := int(t.Weekday())
	if week == 0 {
		week = 7
	}
	return week
}

//获取年
func Year(t time.Time) int {
	t = GetZero(t)
	return t.Year()
}

//获取月
func Month(t time.Time) int {
	t = GetZero(t)
	return int(t.Month())
}

//获取日
func Day(t time.Time) int {
	t = GetZero(t)
	return t.Day()
}
