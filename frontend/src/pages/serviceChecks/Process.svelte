<script lang="ts" module>
  import { get } from 'svelte/store'
  import T, { _ } from '../../includes/Translate.svelte'

  export const validator = (id: string, value: any): string => {
    if (id === 'minimum' && (!value || value < 0))
      return get(_)('ServiceChecks.process.minimum.required')
    if (id === 'value' && !value) return get(_)('ServiceChecks.process.value.required')

    return ''
  }
</script>

<script lang="ts">
  import { Col } from '@sveltestrap/sveltestrap'
  import type { ServiceConfig } from '../../api/notifiarrConfig'
  import type { ChildProps } from '../../includes/Instances.svelte'
  import Input from '../../includes/Input.svelte'
  import CheckedInput from '../../includes/CheckedInput.svelte'

  let {
    form = $bindable(),
    original,
    app,
    index,
    validate,
  }: ChildProps<ServiceConfig> = $props()

  const setData = (expect: string) => {
    return {
      running: expect === 'running',
      restarts: expect.split(',').includes('restart') ?? false,
      minimum: Number(expect.split(':')[1] ?? 1),
      maximum: Number(expect.split(':')[2]?.split(',')[0] ?? 0),
    }
  }

  const originalProc = $derived(setData(original?.expect ?? ''))
  let procCheck = $state(setData(form?.expect ?? ''))
  export const reset = () => (procCheck = setData(form?.expect ?? ''))

  // Input calls validate from `$derived`. Writing form.expect there throws
  // state_unsafe_mutation, so persist the count string from an effect.
  $effect(() => {
    if (!form) return
    if (procCheck.running) {
      // Assign without reading these fields so the effect does not depend on them.
      procCheck.restarts = false
      procCheck.minimum = 1
      procCheck.maximum = 0
      if (form.expect !== 'running') form.expect = 'running'
      return
    }
    const os = original?.expect?.split(':').length ?? 0
    const next =
      `count:${procCheck.minimum}` +
      (procCheck.maximum > 0 || os > 2 ? `:${procCheck.maximum}` : '') +
      (procCheck.restarts ? ',restart' : '')
    if (form.expect !== next) form.expect = next
  })
</script>

<Col md={12}>
  <CheckedInput
    id="value"
    envVar={app.envPrefix + '_' + index + '_VALUE'}
    app={{ ...app, name: 'process' }}
    {index}
    bind:form
    bind:original
    {validate}
    label={$_(app.id + '.process.value.label')}
    description={$_(app.id + '.process.value.description')}
    tooltip={$_(app.id + '.process.value.tooltip')}
    placeholder={$_(app.id + '.process.value.placeholder')} />
</Col>

<Col md={6}>
  <Input
    type="select"
    id={app.id + '.process.running'}
    envVar={`${app.envPrefix}_${index}_EXPECT`}
    bind:value={procCheck.running}
    original={originalProc.running}
    {validate} />
</Col>

<Col md={6}>
  <Input
    disabled={procCheck.running}
    type="select"
    id={app.id + '.process.restarts'}
    envVar={`${app.envPrefix}_${index}_EXPECT`}
    bind:value={procCheck.restarts}
    original={originalProc.restarts}
    {validate} />
</Col>

<Col md={6}>
  <Input
    disabled={procCheck.running}
    type="number"
    min={1}
    id={app.id + '.process.minimum'}
    envVar={`${app.envPrefix}_${index}_EXPECT`}
    bind:value={procCheck.minimum}
    original={originalProc.minimum}
    {validate} />
</Col>

<Col md={6}>
  <Input
    disabled={procCheck.running}
    type="number"
    id={app.id + '.process.maximum'}
    envVar={`${app.envPrefix}_${index}_EXPECT`}
    bind:value={procCheck.maximum}
    original={originalProc.maximum}
    {validate} />
</Col>
