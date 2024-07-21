package mnd

import (
	"time"

	"golift.io/version"
)

const (
	leapDay    = 60 // day of year leap day falls on.
	altLeapDay = 366
)

func today(when time.Time) int {
	switch today := when.YearDay(); {
	case !isLeapYear(when.Year()), today < leapDay:
		return today
	case today == leapDay:
		return altLeapDay
	default:
		return today - 1
	}
}

func isLeapYear(year int) bool {
	return year%400 == 0 || (year%4 == 0 && year%100 != 0)
}

func emojiMonth(when time.Time) string {
	return map[time.Month]string{
		time.January:   "🤖", //
		time.February:  "😻", //
		time.March:     "🗼", //
		time.April:     "🌧", //
		time.May:       "🌸", //
		time.June:      "🍄", //
		time.July:      "🌵", //
		time.August:    "🔥", //
		time.September: "🐸", //
		time.October:   "🍁", //
		time.November:  "👽", //
		time.December:  "⛄", //
	}[when.Month()]
}

// TodaysEmoji returns an emoji specific to the month (or perhaps date).
func TodaysEmoji() string {
	if emoji, exists := map[int]string{
		1:          "🎉", // January 1
		45:         "💝", // February 14
		185:        "🧨", // July 4
		229:        "🏄", // August 17
		254:        "⛑", // September 11
		304:        "🎃", // October 31
		315:        "🪖", // November 11
		328:        "🦃", // November 24
		359:        "🎄", // December 25
		altLeapDay: "🤹", // February 29 (Leap Day)
	}[today(version.Started)]; exists {
		return emoji
	}

	return emojiMonth(version.Started)
}
