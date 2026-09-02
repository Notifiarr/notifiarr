import { fireEvent, render, screen } from '@testing-library/svelte'
import { tick } from 'svelte'
import { describe, expect, it, vi } from 'vitest'
import type { Config } from '../api/notifiarrConfig'
import { FormListTracker, type App } from './formsTracker.svelte'
import Instance from './Instance.svelte'

vi.mock('../api/profile.svelte', async () => {
  const { writable } = await import('svelte/store')
  return { profile: writable({ environment: {} as Record<string, string> }) }
})

type Nvidia = { smiPath: string; busIDs: string[]; disabled: boolean }

const nvidiaApp: App<Nvidia> = {
  id: 'SnapshotApps.Nvidia',
  name: 'Nvidia',
  logo: '',
  envPrefix: 'SNAPSHOT_NVIDIA',
  hidden: ['deletes'],
  empty: { busIDs: [''], smiPath: '', disabled: false },
  merge: () => ({}) as Config,
}

describe('Instance Nvidia bus IDs', () => {
  it('writes textarea edits onto form.busIDs', async () => {
    const flt = new FormListTracker(
      [{ smiPath: '', busIDs: ['GPU-old'], disabled: false }],
      nvidiaApp,
    )
    const form = flt.instances[0]

    render(Instance, {
      props: {
        form,
        original: { smiPath: '', busIDs: ['GPU-old'], disabled: false },
        app: nvidiaApp,
        index: 0,
        indexed: false,
        validate: (id: string, value: string) => flt.validate(id, value, 0),
      },
    })

    await tick()
    const box = screen.getByLabelText('Bus IDs')
    await fireEvent.input(box, { target: { value: 'GPU-new-1\nGPU-new-2' } })
    await tick()
    await tick()

    expect(form.busIDs).toEqual(['GPU-new-1', 'GPU-new-2'])
    expect(flt.formChanged).toBe(true)
  })

  it('reseeds the textarea when the form object is replaced', async () => {
    const first = { smiPath: '', busIDs: ['GPU-old'], disabled: false }
    const flt = new FormListTracker([first], nvidiaApp)

    const { rerender } = render(Instance, {
      props: {
        form: flt.instances[0],
        original: { ...first },
        app: nvidiaApp,
        index: 0,
        indexed: false,
        validate: (id: string, value: string) => flt.validate(id, value, 0),
      },
    })

    await tick()
    await fireEvent.input(screen.getByLabelText('Bus IDs'), {
      target: { value: 'GPU-typed' },
    })
    await tick()

    const reloaded = { smiPath: '', busIDs: ['GPU-from-server'], disabled: false }
    flt.instances = [reloaded]
    await rerender({
      form: flt.instances[0],
      original: { ...reloaded },
      app: nvidiaApp,
      index: 0,
      indexed: false,
      validate: (id: string, value: string) => flt.validate(id, value, 0),
    })
    await tick()
    await tick()

    expect(flt.instances[0].busIDs).toEqual(['GPU-from-server'])
    expect((screen.getByLabelText('Bus IDs') as HTMLTextAreaElement).value).toBe(
      'GPU-from-server',
    )
  })
})
