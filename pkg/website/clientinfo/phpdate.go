package clientinfo

import (
	"strings"
	"time"
)

// PHPDate allows us to easily convert a PHP date format in Go.
type PHPDate struct {
	php string
	fmt string
}

// UnmarshalJSON turns a php date/time format into a golang struct.
func (p *PHPDate) UnmarshalJSON(b []byte) error {
	p.php = (strings.Trim(string(b), `"`))
	for _, character := range p.php {
		p.fmt += flip(string(character))
	}

	return nil
}

// String returns the golang date format for a php date/time format.
func (p *PHPDate) String() string {
	return p.fmt
}

// PHP returns the format for a php date/time.
func (p *PHPDate) PHP() string {
	return p.php
}

// Format returns the date, formatted.
func (p *PHPDate) Format(date time.Time) string {
	return date.Format(p.fmt)
}

// flip the character in the php date/time format string to a go format string.
func flip(character string) string { //nolint:funlen,cyclop
	switch character {
	case "d": // Day of the month, 2 digits with leading zeros
		return "02"
	case "D": // A textual representation of a day, three letters
		return "Mon"
	case "j": // Day of the month without leading zeros
		return "2"
	// case "l": // A full textual representation of the day of the week
	// case "w": // Numeric representation of the day of the week
	// case "z": // The day of the year (starting from 0)
	// case "W": // ISO-8601 week number of year, weeks starting on Monday
	case "F": // A full textual representation of a month
		return "January"
	case "m": // Numeric representation of a month, with leading zeros
		return "01"
	case "M": // A short textual representation of a month, three letters
		return "Jan"
	case "n": // Numeric representation of a month, without leading zeros
		return "1"
	// case "t": // Number of days in the given month
	// case "L": // Whether it's a leap year (1/0)
	case "o":
		fallthrough
	case "Y": // A full numeric representation of a year, 4 digits
		return "2006"
	case "y": // A two digit representation of a year
		return "06"
	case "a": // Lowercase Ante meridiem and Post meridiem
		return "pm"
	case "A": // Uppercase Ante meridiem and Post meridiem
		return "PM"
	case "g": // 12-hour format of an hour without leading zeros
		return "3"
	case "G": // 24-hour format of an hour without leading zeros
		return "15"
	case "h": // 12-hour format of an hour with leading zeros
		return "03"
	case "H": // 24-hour format of an hour with leading zeros
		return "15"
	case "i": // Minutes with leading zeros
		return "04"
	case "s": // Seconds, with leading zeros
		return "05"
	case "e": // Timezone identifier
		fallthrough
	case "T": // Timezone abbreviation
		return "MST"
	case "O": // Difference to Greenwich time (GMT) in hours
		return "-0700"
	case "P": // Difference to Greenwich time (GMT) with colon between hours and minutes
		return "-07:00"
	// case "U": // Seconds since the Unix Epoch (January 1 1970 00:00:00 GMT)
	default:
		return character
	}
}
