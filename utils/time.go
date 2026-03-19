package utils

import "time"

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func BeginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
