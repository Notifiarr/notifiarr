import { describe, expect, it } from 'vitest'
import {
  Frequency,
  Weekday,
  type CronJob,
  type Endpoint,
} from '../../api/notifiarrConfig'
import { FormListTracker } from '../../includes/formsTracker.svelte'
import { validator as cronValidator } from './CronScheduler.svelte'
import { app } from './page.svelte'
import { cronDesc, cronFieldVisible, cronTimes } from './schedule'

const weekly = (daysOfWeek: number[]): Endpoint => ({
  name: 'weekly',
  url: 'https://example.com',
  method: 'GET',
  body: '',
  template: 'ok',
  follow: false,
  frequency: Frequency.Weekly,
  interval: 0,
  timeout: '0s',
  validSsl: true,
  query: null,
  header: null,
  atTimes: [[0, 0, 0]],
  daysOfWeek,
  daysOfMonth: [],
  months: null,
})

describe('cronFieldVisible', () => {
  it('hides every cron field when the schedule is disabled', () => {
    expect(cronFieldVisible('atTimes', Frequency.DeadCron)).toBe(false)
    expect(cronFieldVisible('daysOfWeek', Frequency.DeadCron)).toBe(false)
    expect(cronFieldVisible('daysOfMonth', Frequency.DeadCron)).toBe(false)
  })

  it('shows only atTimes for daily', () => {
    expect(cronFieldVisible('atTimes', Frequency.Daily)).toBe(true)
    expect(cronFieldVisible('daysOfWeek', Frequency.Daily)).toBe(false)
    expect(cronFieldVisible('daysOfMonth', Frequency.Daily)).toBe(false)
  })

  it('shows atTimes and daysOfWeek for weekly, not daysOfMonth', () => {
    expect(cronFieldVisible('atTimes', Frequency.Weekly)).toBe(true)
    expect(cronFieldVisible('daysOfWeek', Frequency.Weekly)).toBe(true)
    expect(cronFieldVisible('daysOfMonth', Frequency.Weekly)).toBe(false)
  })
})

describe('cronValidator', () => {
  it('rejects empty day lists the same way it rejects a zero count', () => {
    expect(cronValidator('daysOfWeek', [])).not.toBe('')
    expect(cronValidator('daysOfMonth', [])).not.toBe('')
    expect(cronValidator('daysOfWeek', 0)).not.toBe('')
    expect(cronValidator('daysOfWeek', [1])).toBe('')
    expect(cronValidator('daysOfWeek', 2)).toBe('')
  })
})

const cron = (partial: Partial<CronJob> & Pick<CronJob, 'frequency'>): CronJob => ({
  interval: 1,
  atTimes: [[9, 5, 0]],
  daysOfWeek: null,
  daysOfMonth: null,
  months: null,
  ...partial,
})

describe('cronTimes', () => {
  it('pads the fields each frequency displays', () => {
    expect(
      cronTimes(
        cron({
          frequency: Frequency.Minutely,
          atTimes: [
            [0, 0, 5],
            [0, 0, 15],
          ],
        }),
      ),
    ).toEqual(['05', '15'])
    expect(
      cronTimes(cron({ frequency: Frequency.Hourly, atTimes: [[0, 7, 5]] })),
    ).toEqual(['07:05'])
    expect(cronTimes(cron({ frequency: Frequency.Daily }))).toEqual(['09:05:00'])
    expect(cronTimes(cron({ frequency: Frequency.Weekly }))).toEqual(['09:05:00'])
    expect(cronTimes(cron({ frequency: Frequency.Monthly }))).toEqual(['09:05:00'])
  })
})

describe('cronDesc', () => {
  it('describes a minutely schedule', () => {
    expect(
      cronDesc(
        cron({
          frequency: Frequency.Minutely,
          atTimes: [
            [0, 0, 5],
            [0, 0, 15],
          ],
        }),
      ),
    ).toBe('2 times every minute at seconds 05, 15')
  })

  it('describes an hourly schedule', () => {
    expect(cronDesc(cron({ frequency: Frequency.Hourly, atTimes: [[0, 7, 5]] }))).toBe(
      'One time every hour at minute:seconds 07:05',
    )
  })

  it('describes a daily schedule', () => {
    expect(cronDesc(cron({ frequency: Frequency.Daily }))).toBe(
      'One time every day at times 09:05:00',
    )
  })

  it('describes a weekly schedule with localized weekdays', () => {
    expect(
      cronDesc(
        cron({
          frequency: Frequency.Weekly,
          daysOfWeek: [Weekday.Monday, Weekday.Friday],
        }),
      ),
    ).toBe('2 times every week on Monday, Friday at times 09:05:00')
  })

  it('describes a monthly schedule', () => {
    expect(cronDesc(cron({ frequency: Frequency.Monthly, daysOfMonth: [1, 15] }))).toBe(
      '2 times every month on days 1, 15 at times 09:05:00',
    )
  })
})

describe('endpoint isValid', () => {
  it('treats a weekly endpoint with no days as invalid', () => {
    const flt = new FormListTracker([weekly([])], app)
    expect(flt.isValid(0)).toBe(false)
    expect(flt.invalid).toBe(true)
  })

  it('treats a weekly endpoint with days as valid', () => {
    const flt = new FormListTracker([weekly([1])], app)
    expect(flt.isValid(0)).toBe(true)
    expect(flt.invalid).toBe(false)
  })
})
