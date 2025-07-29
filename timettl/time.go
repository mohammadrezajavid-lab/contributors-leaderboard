package timettl

import (
	"fmt"
	"time"
)

func GetTTLToEndOfYear() time.Duration {
	now := time.Now()

	firstOfNextYear := time.Date(
		now.Year()+1, time.January, 1, 0, 0, 0, 0, now.Location(),
	)

	return time.Until(firstOfNextYear)
}

func GetTTLToEndOfMonth() time.Duration {
	now := time.Now()

	firstOfNextMonth := time.Date(
		now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location(),
	)

	return time.Until(firstOfNextMonth)
}

func GetTTLToEndOfWeek() time.Duration {
	now := time.Now()

	daysUntilEndOfWeek := (7 - int(now.Weekday())) % 7
	if daysUntilEndOfWeek == 0 {
		daysUntilEndOfWeek = 7
	}

	endOfWeek := time.Date(
		now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location(),
	).AddDate(0, 0, daysUntilEndOfWeek)

	return time.Until(endOfWeek)
}

func GetYear() string {
	now := time.Now()
	return fmt.Sprintf("%d", now.Year())
}

func GetMonth() string {
	now := time.Now()
	return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
}

func GetWeek() string {
	now := time.Now()
	year, week := now.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}
