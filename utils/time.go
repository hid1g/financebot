package utils

import "time"

func CurrentMonth() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := now

	return start, end
}

func LastMonth() (time.Time, time.Time) {

	now := time.Now()
	currentMonthStart := time.Date(
		now.Year(),
		now.Month(),
		1,
		0, 0, 0, 0,
		now.Location(),
	)
	prevMonthStart := currentMonthStart.AddDate(0, -1, 0)
	return prevMonthStart, currentMonthStart

}
