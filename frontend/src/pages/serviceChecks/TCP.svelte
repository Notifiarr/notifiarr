<script lang="ts" module>
  import { get } from 'svelte/store'
  import T, { _ } from '../../includes/Translate.svelte'

  export const validator = (id: string, value: any): string => {
    if (id === 'value' && (!value || !value.includes(':')))
      return get(_)('ServiceChecks.tcp.value.required')

    return ''
  }
</script>

<script lang="ts">
  import { Col } from '@sveltestrap/sveltestrap'
  import type { ServiceConfig } from '../../api/notifiarrConfig'
  import type { ChildProps } from '../../includes/Instances.svelte'
  import CheckedInput from '../../includes/CheckedInput.svelte'
  import { onMount } from 'svelte'

  let {
    form = $bindable(),
    original,
    app,
    index,
    validate,
  }: ChildProps<ServiceConfig> = $props()

  onMount(() => {
    validate?.(app.id + '.value', form.value)
    return () => validate?.(app.id + '.value', 'this.is.valid')
  })
</script>

<Col md={12}>
  <CheckedInput
    id="value"
    envVar={`${app.envPrefix}_${index}_VALUE`}
    app={{ ...app, name: 'tcp' }}
    {index}
    bind:form
    bind:original
    {validate}
    label={$_(app.id + '.tcp.value.label')}
    description={$_(app.id + '.tcp.value.description')}
    tooltip={$_(app.id + '.tcp.value.tooltip')}
    placeholder={$_(app.id + '.tcp.value.placeholder')} />
</Col>
