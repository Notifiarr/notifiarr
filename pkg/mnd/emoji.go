package mnd

import (
	"time"

	"golift.io/version"
)

const (
	leapDay    = 60
	altLeapDay = 366
)

// TodaysEmoji returns an emoji specific to the month (or perhaps date).
func TodaysEmoji() string {
	today := version.Started.YearDay()

	switch year := version.Started.Year(); {
	case !leapYear(year), today < leapDay:
		break
	case today == leapDay:
		today = altLeapDay
	default:
		today--
	}

	if emoji, ok := specialDays[today]; ok {
		return emoji
	}

	return monthEmojis[version.Started.Month()]
}

func leapYear(year int) bool {
	return year%400 == 0 || (year%4 == 0 && year%100 != 0)
}

var monthEmojis = map[time.Month]string{ //nolint:gochecknoglobals
	time.January:   "🤖", //
	time.February:  "😻", //
	time.March:     "🗼", //
	time.April:     "🌦", //
	time.May:       "🌸", //
	time.June:      "🍀", //
	time.July:      "🌵", //
	time.August:    "🔥", //
	time.September: "🍁", //
	time.October:   "🍉", //
	time.November:  "🍗", //
	time.December:  "⛄", //
}

var specialDays = map[int]string{ //nolint:gochecknoglobals
	1:          "🎉", // January 1
	45:         "💝", // February 14
	185:        "🧨", // July 4
	229:        "🏄", // August 17
	304:        "🎃", // October 31
	315:        "🪖", // November 11
	328:        "🦃", // November 24
	359:        "🎄", // December 25
	altLeapDay: "🤹", // February 29 (Leap Day)
}
