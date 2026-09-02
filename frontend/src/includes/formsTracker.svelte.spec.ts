import { describe, expect, it } from 'vitest'
import type { Config } from '../api/notifiarrConfig'
import { FormListTracker, type App } from './formsTracker.svelte'

type Row = { name: string; url: string }

const app: App<Row> = {
  id: 'Test.App',
  name: 'Test',
  logo: '',
  envPrefix: 'TEST',
  empty: { name: '', url: '' },
  merge: () => ({}) as Config,
  validator: (id, value) => {
    if (id.endsWith('.name') && !value) return 'name required'
    return ''
  },
}

describe('FormListTracker', () => {
  it('reports invalid from instance values without writing a feedback map', () => {
    const flt = new FormListTracker([{ name: '', url: 'http://x' }], app)
    expect(flt.validate('Test.App.name', '', 0)).toBe('name required')
    expect(flt.isValid(0)).toBe(false)
    expect(flt.invalid).toBe(true)
  })

  it('is valid when every field passes', () => {
    const flt = new FormListTracker([{ name: 'ok', url: 'http://x' }], app)
    expect(flt.validate('Test.App.name', 'ok', 0)).toBe('')
    expect(flt.isValid(0)).toBe(true)
    expect(flt.invalid).toBe(false)
  })

  it('recomputes invalid after resetAll', () => {
    const flt = new FormListTracker([{ name: 'ok', url: 'http://x' }], app)
    flt.instances[0].name = ''
    expect(flt.invalid).toBe(true)
    flt.resetAll()
    expect(flt.instances[0].name).toBe('ok')
    expect(flt.invalid).toBe(false)
  })
})
