package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// Frequency sets the base "how-often" a CronJob is executed.
type Frequency int

// Frequency values that are known and work.
const (
	Custom Frequency = iota
	Minutely
	Hourly
	Daily
	Weekly
	Monthly
)

// AtTimes is a list of times in: hours, minutes, seconds.
type AtTimes [][3]uint

// CronJob defines when a job should run.
// When Frequency is set to:
// `Custom` uses all the fields.
// `Minutely` uses Seconds.
// `Hourly` uses Minutes and Seconds.
// `Daily` uses Hours, Minutes and Seconds.
// `Weekly` uses DaysOfWeek, Hours, Minutes and Seconds.
// `Monthly` uses DaysOfMonth, Hours, Minutes and Seconds.
type CronJob struct {
	Frequency
	AtTimes     AtTimes
	DaysOfWeek  []time.Weekday // 0-6, sunday = 0
	DaysOfMonth []int          // gocron.days
	Months      []uint         // 1-12.
}

func (c *CronJob) DaysOfTheMonths() gocron.DaysOfTheMonth {
	mon1, mon2 := func() (int, []int) {
		switch len(c.DaysOfMonth) {
		case 0:
			return 0, nil
		case 1:
			return c.DaysOfMonth[0], nil
		default:
			return c.DaysOfMonth[0], c.DaysOfMonth[1:]
		}
	}()

	return gocron.NewDaysOfTheMonth(mon1, mon2...)
}

func (c *CronJob) DaysOfTheWeek() func() []time.Weekday {
	return func() []time.Weekday { return c.DaysOfWeek }
}

func (a AtTimes) AtTimes() func() []gocron.AtTime {
	output := []gocron.AtTime{}

	const (
		maxHours  = 23
		maxMinSec = 59
	)

	for _, times := range a {
		if times[0] > maxHours {
			times[0] = 0
		}

		if times[1] > maxMinSec {
			times[1] = 0
		}

		if times[2] > maxMinSec {
			times[2] = 0
		}

		output = append(output, gocron.NewAtTime(times[0], times[1], times[2]))
	}

	return func() []gocron.AtTime { return output }
}

func (a AtTimes) Hours() []uint {
	return a.getField(0)
}

func (a AtTimes) Minutes() []uint {
	return a.getField(1)
}

func (a AtTimes) Seconds() []uint {
	return a.getField(2)
}

func (a AtTimes) getField(field int) []uint {
	if len(a) == 0 {
		return []uint{0}
	}

	output := make([]uint, 0, len(a))

	for idx := range a {
		output = append(output, a[idx][field])
	}
	return output
}

func JoinOrStar[T []int | []uint | []time.Weekday](input T) string {
	if len(input) == 0 {
		return "*"
	}

	return strings.Trim(strings.ReplaceAll(fmt.Sprint(input), " ", ","), "[]")
}

func JoinOrZero[T []int | []uint | []time.Weekday](input T) string {
	if len(input) == 0 {
		return "0"
	}

	return strings.Trim(strings.ReplaceAll(fmt.Sprint(input), " ", ","), "[]")
}
