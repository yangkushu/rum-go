package utils

import "time"

// TimeDayStart 获取某天开始的时间
func TimeDayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// TimeDayEnd 获取某天结束的时间
func TimeDayEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

// TimeTodayStart 获取今天开始的时间
func TimeTodayStart() time.Time {
	return TimeDayStart(time.Now())
}

// TimeTodayEnd 获取今天结束的时间
func TimeTodayEnd() time.Time {
	return TimeDayEnd(time.Now())
}

// TimeIsToday 时间是今天
func TimeIsToday(t time.Time) bool {
	return TimeDayStart(t).Equal(TimeTodayStart())
}
