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
  import { onMount } from 'svelte'

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

  const originalProc = $derived(setData(original.expect))
  let procCheck = $state(setData(form.expect))
  export const reset = () => (procCheck = setData(form.expect))

  const updateExpect = (id: string, value: any) => {
    if (procCheck.running) {
      form.expect = 'running'
      procCheck.restarts = false
      procCheck.minimum = 1
      procCheck.maximum = 0
      resetValidators()
    } else {
      const os = original.expect.split(':').length
      form.expect =
        `count:${procCheck.minimum}` +
        (procCheck.maximum > 0 || os > 2 ? `:${procCheck.maximum}` : '') +
        (procCheck.restarts ? ',restart' : '')
      validate?.(app.id + '.process.restarts', procCheck.restarts)
      validate?.(app.id + '.process.minimum', procCheck.minimum)
      validate?.(app.id + '.process.maximum', procCheck.maximum)
    }

    return id ? (validate?.(id, value) ?? '') : ''
  }

  const resetValidators = () => {
    validate?.(app.id + '.value', 'this.is.valid')
    validate?.(app.id + '.process.restarts', 'false')
    validate?.(app.id + '.process.minimum', 1)
    validate?.(app.id + '.process.maximum', 0)
  }

  onMount(() => {
    updateExpect('', '')
    return () => resetValidators()
  })
</script>

<Col md={12}>
  <CheckedInput
    id="value"
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
    bind:value={procCheck.running}
    original={originalProc.running}
    validate={updateExpect} />
</Col>

<Col md={6}>
  <Input
    disabled={procCheck.running}
    type="select"
    id={app.id + '.process.restarts'}
    bind:value={procCheck.restarts}
    original={originalProc.restarts}
    validate={updateExpect} />
</Col>

<Col md={6}>
  <Input
    disabled={procCheck.running}
    type="number"
    min={1}
    id={app.id + '.process.minimum'}
    bind:value={procCheck.minimum}
    original={originalProc.minimum}
    validate={updateExpect} />
</Col>

<Col md={6}>
  <Input
    disabled={procCheck.running}
    type="number"
    id={app.id + '.process.maximum'}
    bind:value={procCheck.maximum}
    original={originalProc.maximum}
    validate={updateExpect} />
</Col>
