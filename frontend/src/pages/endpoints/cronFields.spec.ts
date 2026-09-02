import { describe, expect, it } from 'vitest'
import { Frequency } from '../../api/notifiarrConfig'
import { cronFieldVisible } from './schedule'

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
