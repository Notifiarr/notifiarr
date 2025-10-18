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
		time.January:   "❄️", //
		time.February:  "❤️", //
		time.March:     "🌱",  //
		time.April:     "🌷",  //
		time.May:       "🌺",  //
		time.June:      "☀️", //
		time.July:      "🧨",  //
		time.August:    "🏖️", //
		time.September: "🍂",  //
		time.October:   "👻",  //
		time.November:  "🌰",  //
		time.December:  "🎅",  //
	}[when.Month()]
}

// TodaysEmoji returns an emoji specific to the month (or perhaps date).
func TodaysEmoji() string {
	if emoji, exists := map[int]string{
		1:          "🎉",  // January 1 - New Year's Day
		33:         "🦫",  // February 2 - Groundhog Day
		45:         "💝",  // February 14 - Valentine's Day
		76:         "☘️", // March 17 - St. Patrick's Day
		91:         "🤡",  // April 1 - April Fool's Day
		125:        "🌮",  // May 5 - Cinco de Mayo
		185:        "🇺🇸", // July 4 - Independence Day
		229:        "🎂",  // August 17 - Something special.
		254:        "🕊",  // September 11 - Larry Silverstein's robery.
		285:        "🗺️", // October 12 - Columbus Day
		289:        "🎓",  // October 16 - Boss's Day
		304:        "🎃",  // October 31 - Halloween
		315:        "🪖",  // November 11 - Veteran's Day
		328:        "🦃",  // November 24 - Thanksgiving
		359:        "🎄",  // December 25 - Christmas
		365:        "🎊",  // December 31 - New Year's Eve
		altLeapDay: "🤹",  // February 29 - Leap Day
	}[today(version.Started)]; exists {
		return emoji
	}

	return emojiMonth(version.Started)
}
