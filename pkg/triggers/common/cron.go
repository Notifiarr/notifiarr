package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/Notifiarr/notifiarr/pkg/website"
	"github.com/go-co-op/gocron/v2"
)

// Frequency sets the base "how-often" a CronJob is executed.
type Frequency int

// Frequency values that are known and work.
const (
	DeadCron Frequency = iota
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
// 0 `DeadCron` disables the schedule.
// 1 `Minutely` uses Seconds.
// 2 `Hourly` uses Minutes and Seconds.
// 3 `Daily` uses Hours, Minutes and Seconds.
// 4 `Weekly` uses DaysOfWeek, Hours, Minutes and Seconds.
// 5 `Monthly` uses DaysOfMonth, Hours, Minutes and Seconds.
type CronJob struct {
	// Frequency to configure the job. Pass 0 disable the cron.
	Frequency Frequency `json:"frequency" toml:"frequency" xml:"frequency" yaml:"frequency"`
	// Interval for Daily, Weekly and Monthly Frequencies. 1 = every day/week/month, 2 = every other, and so on.
	Interval int `json:"interval" toml:"interval" xml:"interval" yaml:"interval"`
	// AtTimes is a list of 'hours, minutes, seconds' to schedule for Daily/Weekly/Monthly frequencies.
	// Also used in Minutely and Hourly schedules, a bit awkwardly.
	AtTimes AtTimes `json:"atTimes" toml:"at_times" xml:"at_times" yaml:"atTimes"`
	// DaysOfWeek is a list of days to schedule. 0-6. 0 = Sunday.
	DaysOfWeek []time.Weekday `json:"daysOfWeek" toml:"days_of_week" xml:"days_of_week" yaml:"daysOfWeek"`
	// DaysOfMonth is a list of days to schedule. 1 to 31 or -31 to -1 to count backward.
	DaysOfMonth []int `json:"daysOfMonth" toml:"days_of_month" xml:"days_of_month" yaml:"daysOfMonth"`
	// Months to schedule. 1 to 12. 1 = January.
	Months []uint `json:"months" toml:"months" xml:"months" yaml:"months"`
}

func (c *CronJob) fix() { //nolint:cyclop
	if c.Interval == 0 {
		c.Interval = 1
	}

	for idx, m := range c.Months {
		if m > maxMonths || m < 1 {
			c.Months[idx] = 1
		}
	}

	for idx, d := range c.DaysOfWeek {
		if d > time.Saturday || d < time.Sunday {
			c.DaysOfWeek[idx] = time.Sunday
		}
	}

	for idx, d := range c.DaysOfMonth {
		if d > maxDays || d < 1 {
			c.DaysOfMonth[idx] = 1
		}
	}
}

func (a AtTimes) AtTimes() func() []gocron.AtTime {
	output := []gocron.AtTime{}

	for _, times := range a {
		if times[fieldHours] > maxHours {
			times[fieldHours] = maxHours
		}

		if times[fieldMinutes] > maxMinSec {
			times[fieldMinutes] = maxMinSec
		}

		if times[fieldSeconds] > maxMinSec {
			times[fieldSeconds] = maxMinSec
		}

		output = append(output, gocron.NewAtTime(times[fieldHours], times[fieldMinutes], times[fieldSeconds]))
	}

	return func() []gocron.AtTime { return output }
}

func (a AtTimes) minutes() string {
	return joinOrZero(a.getField(fieldMinutes))
}

func (a AtTimes) seconds() string {
	return joinOrZero(a.getField(fieldSeconds))
}

const (
	maxDays   = 31
	maxHours  = 23
	maxMinSec = 59
	maxMonths = 12
)

type field int

const (
	fieldHours field = iota
	fieldMinutes
	fieldSeconds
)

func (a AtTimes) getField(field field) []uint {
	if len(a) == 0 || field > fieldSeconds || field < 0 {
		return nil
	}

	output := make([]uint, 0, len(a))

	for idx := range a {
		output = append(output, a[idx][field])
	}

	return output
}

func joinOrStar[T []int | []uint | []time.Weekday](input T) string {
	if len(input) == 0 {
		return "*"
	}

	return strings.Trim(strings.ReplaceAll(fmt.Sprint(input), " ", ","), "[]")
}

func joinOrZero(input []uint) string {
	if len(input) == 0 {
		return "0"
	}

	return strings.Trim(strings.ReplaceAll(fmt.Sprint(input), " ", ","), "[]")
}

func (c *CronJob) daysOfTheMonths() gocron.DaysOfTheMonth {
	switch len(c.DaysOfMonth) {
	case 0:
		return gocron.NewDaysOfTheMonth(1)
	case 1:
		return gocron.NewDaysOfTheMonth(c.DaysOfMonth[0])
	default:
		return gocron.NewDaysOfTheMonth(c.DaysOfMonth[0], c.DaysOfMonth[1:]...)
	}
}

func (c *CronJob) daysOfTheWeek() func() []time.Weekday {
	return func() []time.Weekday { return c.DaysOfWeek }
}

func (a *Action) newCron(cron gocron.Scheduler) {
	var def gocron.JobDefinition
	switch a.J.fix(); a.J.Frequency {
	case DeadCron:
		return
	case Minutely:
		def = gocron.CronJob(a.J.AtTimes.seconds()+" * * * * *", true)
	case Hourly:
		def = gocron.CronJob(a.J.AtTimes.seconds()+" "+a.J.AtTimes.minutes()+" * * * *", true)
	case Daily:
		def = gocron.DailyJob(1, a.J.AtTimes.AtTimes())
	case Weekly:
		def = gocron.WeeklyJob(1, a.J.daysOfTheWeek(), a.J.AtTimes.AtTimes())
	case Monthly:
		def = gocron.MonthlyJob(1, a.J.daysOfTheMonths(), a.J.AtTimes.AtTimes())
	}

	var err error

	a.job, err = cron.NewJob(def, gocron.NewTask(
		func() { a.C <- &ActionInput{Type: website.EventSched} },
	))
	if err != nil {
		panic(fmt.Sprint("THIS IS A BUG, please report it: ", err))
	}
}
