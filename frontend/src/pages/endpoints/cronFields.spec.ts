import { describe, expect, it } from 'vitest'
import { Frequency, type Endpoint } from '../../api/notifiarrConfig'
import { FormListTracker } from '../../includes/formsTracker.svelte'
import { validator as cronValidator } from './CronScheduler.svelte'
import { app } from './page.svelte'
import { cronFieldVisible } from './schedule'

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
