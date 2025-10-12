<script lang="ts" module>
  import { get } from 'svelte/store'
  import T, { _ } from '../../includes/Translate.svelte'

  export const validator = (id: string, value: any): string => {
    if (id === 'count' && (!value || value < 0))
      return get(_)('ServiceChecks.ping.count.required')
    if (id === 'minimum' && (!value || value < 0))
      return get(_)('ServiceChecks.ping.minimum.required')
    if (id === 'value' && !value) return get(_)('ServiceChecks.ping.value.required')

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
      count: Number(expect.split(':')[0] || 1),
      minimum: Number(expect.split(':')[1] || 1),
      interval: Number(expect.split(':')[2] || 10),
    }
  }

  const originalPing = $derived(setData(original.expect))
  let pingCheck = $state(setData(form.expect))
  export const reset = () => (pingCheck = setData(form.expect))

  const updateExpect = (id: string, value: any) => {
    validate?.(app.id + '.ping.count', pingCheck.count)
    validate?.(app.id + '.ping.minimum', pingCheck.minimum)
    validate?.(app.id + '.ping.interval', pingCheck.interval)
    form.expect = `${pingCheck.count || 1}:${pingCheck.minimum || 1}:${pingCheck.interval || 10}`

    return id ? (validate?.(id, value) ?? '') : ''
  }

  const resetValidators = () => {
    validate?.(app.id + '.value', 'this.is.valid')
    validate?.(app.id + '.ping.count', 1)
    validate?.(app.id + '.ping.minimum', 1)
    validate?.(app.id + '.ping.interval', 10)
  }

  onMount(() => {
    updateExpect('', '')
    return () => resetValidators()
  })
</script>

<Col md={12}>
  <CheckedInput
    id="value"
    envVar={`${app.envPrefix}_${index}_VALUE`}
    app={{ ...app, name: 'ping' }}
    {index}
    bind:form
    bind:original
    {validate}
    label={$_(app.id + '.ping.value.label')}
    description={$_(app.id + '.ping.value.description')}
    tooltip={$_(app.id + '.ping.value.tooltip')}
    placeholder={$_(app.id + '.ping.value.placeholder')} />
</Col>

<Col md={4}>
  <Input
    type="number"
    min={1}
    envVar={`${app.envPrefix}_${index}_EXPECT`}
    id={app.id + '.ping.count'}
    bind:value={pingCheck.count}
    original={originalPing.count}
    validate={updateExpect} />
</Col>

<Col md={4}>
  <Input
    type="number"
    min={1}
    envVar={`${app.envPrefix}_${index}_EXPECT`}
    id={app.id + '.ping.minimum'}
    bind:value={pingCheck.minimum}
    original={originalPing.minimum}
    validate={updateExpect} />
</Col>

<Col md={4}>
  <Input
    type="select"
    envVar={`${app.envPrefix}_${index}_EXPECT`}
    id={app.id + '.ping.interval'}
    bind:value={pingCheck.interval}
    original={originalPing.interval}
    options={[10, 20, 30, 50, 100, 200, 300, 500].map(ms => ({
      name: $_('words.clock.short.ms', { values: { milliseconds: ms } }),
      value: ms,
    }))}
    validate={updateExpect} />
</Col>
