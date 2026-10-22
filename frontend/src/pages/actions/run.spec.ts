import { beforeEach, describe, expect, it, vi } from 'vitest'
import { Frequency, type TriggerInfo } from '../../api/notifiarrConfig'
import { profile } from '../../api/profile.svelte'
import { dur, option, val } from './run'

vi.mock('../../api/profile.svelte', async () => {
  const { writable } = await import('svelte/store')
  return {
    profile: writable({
      config: { commands: [] as ({ name: string; hash: string } | null)[] },
      siteCrons: [] as ({ name: string } | null)[],
    }),
  }
})

vi.mock('../../api/fetch', () => ({ getUi: vi.fn() }))

vi.mock('../../header/Reload.svelte', () => ({ reload: vi.fn() }))

type ProfileBits = {
  config: { commands: ({ name: string; hash: string } | null)[] }
  siteCrons: ({ name: string } | null)[]
}

function setProfile(value: ProfileBits) {
  ;(profile as unknown as { set: (v: ProfileBits) => void }).set(value)
}

const row = (
  partial: Partial<TriggerInfo> & Pick<TriggerInfo, 'key' | 'kind' | 'name'>,
): TriggerInfo => ({ runs: 0, ...partial })

describe('option', () => {
  beforeEach(() => {
    setProfile({
      config: { commands: [null, { name: 'backup', hash: 'abc123' }] },
      siteCrons: [null, { name: 'nightly' }, { name: 'hourly' }],
    })
  })

  it('returns the custom command hash and rip when the command is missing', () => {
    expect(
      option(row({ key: 'TrigCustomCommand', kind: 'Trigger', name: 'backup' })),
    ).toBe('abc123')
    expect(
      option(row({ key: 'TrigCustomCommand', kind: 'Trigger', name: 'missing' })),
    ).toBe('rip')
  })

  it('returns the index of a website cron after skipping null holes', () => {
    expect(
      option(
        row({
          key: 'TrigCustomCronTimer',
          kind: 'Schedule',
          name: "Website timer: 'hourly'",
        }),
      ),
    ).toBe('1')
  })

  it('returns the endpoint name and nothing for a plain trigger', () => {
    expect(option(row({ key: 'TrigEndpointURL', kind: 'Trigger', name: 'health' }))).toBe(
      'health',
    )
    expect(option(row({ key: 'TrigStop', kind: 'Trigger', name: 'reload' }))).toBe('')
  })
})

describe('dur', () => {
  it('says Never for an unscheduled trigger', () => {
    expect(dur(row({ key: 'TrigStop', kind: 'Trigger', name: 'reload' }))).toBe('Never')
  })

  it('formats a timer interval with seconds', () => {
    expect(
      dur(
        row({ key: 'TrigSnapshot', kind: 'Timer', name: 'snapshot', interval: 90_000 }),
      ),
    ).toBe('Every 1m 30s')
  })

  it('formats a cron schedule', () => {
    expect(
      dur(
        row({
          key: 'TrigCustomCronTimer',
          kind: 'Schedule',
          name: "sync 'nightly'",
          cron: {
            frequency: Frequency.Daily,
            interval: 1,
            atTimes: [[9, 5, 0]],
            daysOfWeek: null,
            daysOfMonth: null,
            months: null,
          },
        }),
      ),
    ).toBe('One time every day at times 09:05:00')
  })
})

describe('val', () => {
  it('uses the quoted cron name in the website-timer label', () => {
    expect(
      val(row({ key: 'TrigCustomCronTimer', kind: 'Schedule', name: "sync 'nightly'" })),
    ).toBe('Website timer: nightly')
  })

  it('interpolates the endpoint name', () => {
    expect(val(row({ key: 'TrigEndpointURL', kind: 'Trigger', name: 'health' }))).toBe(
      'Endpoint URL relayed to website: health',
    )
  })
})
