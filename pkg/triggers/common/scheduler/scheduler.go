package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// Frequency sets the base "how-often" a CronJob is executed.
// See the Frequency constants.
type Frequency uint

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

// A Weekday specifies a day of the week (Sunday = 0, ...).
// Copied from stdlib to avoid the String method.
type Weekday int

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

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
	Interval uint `json:"interval" toml:"interval" xml:"interval" yaml:"interval"`
	// AtTimes is a list of 'hours, minutes, seconds' to schedule for Daily/Weekly/Monthly frequencies.
	// Also used in Minutely and Hourly schedules, a bit awkwardly.
	AtTimes AtTimes `json:"atTimes" toml:"at_times" xml:"at_times" yaml:"atTimes"`
	// DaysOfWeek is a list of days to schedule. 0-6. 0 = Sunday.
	DaysOfWeek []Weekday `json:"daysOfWeek" toml:"days_of_week" xml:"days_of_week" yaml:"daysOfWeek"`
	// DaysOfMonth is a list of days to schedule. 1 to 31 or -31 to -1 to count backward.
	DaysOfMonth []int `json:"daysOfMonth" toml:"days_of_month" xml:"days_of_month" yaml:"daysOfMonth"`
	// Months to schedule. 1 to 12. 1 = January.
	Months []uint `json:"months" toml:"months" xml:"months" yaml:"months"`
}

// String attempts to turn a CronJob into a string.
func (c *CronJob) String() string {
	switch c.Frequency {
	default:
		fallthrough
	case DeadCron:
		return "Schedule Disabled"
	case Minutely:
		return fmt.Sprintf("Every minute at seconds %v", c.AtTimes.seconds())
	case Hourly:
		return fmt.Sprintf("Every hour at minutes %v", c.AtTimes.minutes())
	case Daily:
		return fmt.Sprintf("Every %d day(s) at times (h:m:s) %v", c.Interval, c.AtTimes)
	case Weekly:
		return fmt.Sprintf("Days of week %v every %d week(s) at times (h:m:s) %v", c.DaysOfWeek, c.Interval, c.AtTimes)
	case Monthly:
		return fmt.Sprintf("Days of month %v every %d month(s) at times (h:m:s) %v", c.DaysOfMonth, c.Interval, c.AtTimes)
	}
}

func (c *CronJob) fix() { //nolint:cyclop
	if c.Interval == 0 {
		c.Interval = 1
	}

	if c.Frequency > Monthly {
		c.Frequency = DeadCron // oops.
	}

	for _, times := range c.AtTimes {
		if times[fieldHours] > maxHours {
			times[fieldHours] = maxHours
		}

		if times[fieldMinutes] > maxMinSec {
			times[fieldMinutes] = maxMinSec
		}

		if times[fieldSeconds] > maxMinSec {
			times[fieldSeconds] = maxMinSec
		}
	}

	for idx, m := range c.Months {
		if m > maxMonths || m < 1 {
			c.Months[idx] = 1
		}
	}

	for idx, d := range c.DaysOfWeek {
		if d > Saturday || d < Sunday {
			c.DaysOfWeek[idx] = Sunday
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
		output = append(output, gocron.NewAtTime(times[fieldHours], times[fieldMinutes], times[fieldSeconds]))
	}

	return func() []gocron.AtTime { return output }
}

func (a AtTimes) minutes() string {
	return joinOr(a.getField(fieldMinutes), "0")
}

func (a AtTimes) seconds() string {
	return joinOr(a.getField(fieldSeconds), "0")
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

func joinOr[T []int | []uint | []time.Weekday](input T, or string) string {
	if len(input) == 0 {
		return or
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
	return func() []time.Weekday {
		days := make([]time.Weekday, 7)
		for _, d := range c.DaysOfWeek {
			days[d] = time.Weekday(d)
		}

		return days
	}
}

func (c *CronJob) New(cron gocron.Scheduler, cmd func()) gocron.Job { //nolint:ireturn,nolintlint // it's what we have.
	var def gocron.JobDefinition
	switch c.fix(); c.Frequency {
	default:
		fallthrough
	case DeadCron:
		return nil
	case Minutely:
		def = gocron.CronJob(c.AtTimes.seconds()+" * * * * *", true)
	case Hourly:
		def = gocron.CronJob(c.AtTimes.minutes()+" * * * *", false)
	case Daily:
		def = gocron.DailyJob(c.Interval, c.AtTimes.AtTimes())
	case Weekly:
		def = gocron.WeeklyJob(c.Interval, c.daysOfTheWeek(), c.AtTimes.AtTimes())
	case Monthly:
		def = gocron.MonthlyJob(c.Interval, c.daysOfTheMonths(), c.AtTimes.AtTimes())
	}

	job, err := cron.NewJob(def, gocron.NewTask(cmd))
	if err != nil {
		panic(fmt.Sprint("[scheduler] THIS IS A BUG, please report it: ", err))
	}

	return job
}
