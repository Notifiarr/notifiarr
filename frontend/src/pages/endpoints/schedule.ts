import { Frequency, Weekday } from '../../api/notifiarrConfig'
import { get } from 'svelte/store'
import type { CronJob } from '../../api/notifiarrConfig'
import { _ } from '../../includes/Translate.svelte'

/** CronScheduler only renders these fields for some frequencies. */
export const cronFieldVisible = (
  id: string,
  frequency: Frequency | undefined,
): boolean => {
  if (frequency == null || frequency === Frequency.DeadCron) return false
  if (id === 'atTimes') return true
  if (id === 'daysOfWeek') return frequency === Frequency.Weekly
  if (id === 'daysOfMonth') return frequency === Frequency.Monthly
  return false
}

/** Pad a number with a leading zero. */
const pad = (n: number) => n.toString().padStart(2, '0')

/** Get a human-readable string of the times. */
export const cronTimes = (cron: CronJob): string[] => {
  if (cron.frequency === Frequency.Minutely)
    return cron.atTimes?.map(t => `${pad(t[2])}`) ?? []
  if (cron.frequency === Frequency.Hourly)
    return cron.atTimes?.map(t => `${pad(t[1])}:${pad(t[2])}`) ?? []
  return cron.atTimes?.map(t => t.map(l => pad(l)).join(':')) ?? []
}

/** The weekdays, localized. Pass `$_` from a component so labels update with locale. */
export const weekdays = (
  t: (id: string) => string = id => String(get(_)(id) ?? id),
): Record<number, string> => ({
  [Weekday.Sunday]: t('words.clock.days.Sunday'),
  [Weekday.Monday]: t('words.clock.days.Monday'),
  [Weekday.Tuesday]: t('words.clock.days.Tuesday'),
  [Weekday.Wednesday]: t('words.clock.days.Wednesday'),
  [Weekday.Thursday]: t('words.clock.days.Thursday'),
  [Weekday.Friday]: t('words.clock.days.Friday'),
  [Weekday.Saturday]: t('words.clock.days.Saturday'),
})

/** Turns a cron job into a human readable, localized string. */
export const cronDesc = (cron: CronJob): string => {
  const times = cronTimes(cron).join(', ') || get(_)('scheduler.times.noneSelected')

  if (cron.frequency === Frequency.Minutely)
    return get(_)('scheduler.times.everyMinute', {
      values: { count: cron.atTimes?.length ?? 1, times },
    })

  if (cron.frequency === Frequency.Hourly)
    return get(_)('scheduler.times.everyHour', {
      values: { count: cron.atTimes?.length ?? 1, times },
    })

  if (cron.frequency === Frequency.Daily)
    return get(_)('scheduler.times.everyDay', {
      values: { count: cron.atTimes?.length ?? 1, times },
    })

  if (cron.frequency === Frequency.Weekly)
    return get(_)('scheduler.times.everyWeek', {
      values: {
        count: (cron.atTimes?.length ?? 1) * (cron.daysOfWeek?.length ?? 1),
        times,
        daysOfWeek:
          cron.daysOfWeek?.map(d => weekdays()[d]).join(', ') ||
          get(_)('scheduler.times.noDaysSelected'),
      },
    })

  if (cron.frequency === Frequency.Monthly)
    return get(_)('scheduler.times.everyMonth', {
      values: {
        count: (cron.atTimes?.length ?? 1) * (cron.daysOfMonth?.length ?? 1),
        times,
        daysOfMonth:
          cron.daysOfMonth?.join(', ') || get(_)('scheduler.times.noDaysSelected'),
      },
    })
  return get(_)('Actions.titles.Never')
}
