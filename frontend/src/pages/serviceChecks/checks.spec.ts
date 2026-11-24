import { fireEvent, render, screen } from '@testing-library/svelte'
import { tick } from 'svelte'
import { describe, expect, it, vi } from 'vitest'
import { CheckType, type Config, type ServiceConfig } from '../../api/notifiarrConfig'
import { FormListTracker, type App } from '../../includes/formsTracker.svelte'
import { empty } from './page.svelte'
import Http from './HTTP.svelte'
import Proc from './Process.svelte'
import Ping from './Ping.svelte'

vi.mock('../../api/profile.svelte', async () => {
  const { writable } = await import('svelte/store')
  return { profile: writable({ environment: {} as Record<string, string> }) }
})

const app: App<ServiceConfig> = {
  id: 'ServiceChecks',
  name: 'Checks',
  logo: '',
  envPrefix: 'SERVICE',
  empty,
  merge: () => ({}) as Config,
  validator: () => '',
}

const check = (over: Partial<ServiceConfig>): ServiceConfig => ({ ...empty, ...over })

describe('service check editors', () => {
  it('keeps http validation pure and syncs the url from an effect', async () => {
    const flt = new FormListTracker(
      [check({ name: 'web', value: 'https://example.com|X-Test: 1', expect: '200' })],
      app,
    )
    const form = flt.instances[0]
    const validate = vi.fn(() => '')

    render(Http, { props: { form, original: flt.original[0], app, index: 0, validate } })
    await tick()

    expect(validate).toHaveBeenCalled()
    expect(form.value).toBe('https://example.com|X-Test: 1')
    expect(form.expect).toBe('200')

    await fireEvent.input(screen.getByLabelText('URL'), {
      target: { value: 'https://example.com/health' },
    })
    await tick()

    expect(form.value).toBe('https://example.com/health|X-Test: 1')
  })

  it('renders a process check without writing state from validation', async () => {
    const flt = new FormListTracker(
      [
        check({
          name: 'proc',
          type: CheckType.process,
          value: 'nginx',
          expect: 'count:2:3',
        }),
      ],
      app,
    )
    const form = flt.instances[0]

    render(Proc, {
      props: { form, original: flt.original[0], app, index: 0, validate: () => '' },
    })
    await tick()

    expect(form.expect).toBe('count:2:3')
    expect(screen.getByLabelText('Process Match')).toBeTruthy()
  })

  it('renders a ping check without rewriting a stable expect string', async () => {
    const flt = new FormListTracker(
      [
        check({
          name: 'ping',
          type: CheckType.ping,
          value: '127.0.0.1',
          expect: '2:1:10',
        }),
      ],
      app,
    )
    const form = flt.instances[0]

    render(Ping, {
      props: { form, original: flt.original[0], app, index: 0, validate: () => '' },
    })
    await tick()

    expect(form.expect).toBe('2:1:10')
  })
})
