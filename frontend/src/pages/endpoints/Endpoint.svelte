<script lang="ts">
  import Input from '../../includes/Input.svelte'
  import { Col, Input as Box, Row, Button } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import { type Endpoint } from '../../api/notifiarrConfig'
  import type { ChildProps } from '../../includes/Instances.svelte'
  import CheckedInput from '../../includes/CheckedInput.svelte'
  import CronScheduler from './CronScheduler.svelte'
  import { mapLength } from '../../includes/util'

  let {
    form = $bindable(),
    original,
    app,
    index = 0, // unused but matches our ChildProps interface.
    validate,
  }: ChildProps<Endpoint> = $props()

  let originalQuery = $derived.by(() =>
    Object.entries(original.query ?? {})
      .filter(([_, value]) => value !== null)
      .map(([key, value]) => value?.map(v => `${key}=${v}`).join('\n') ?? '')
      .join('\n'),
  )

  let originalHeader = $derived.by(() =>
    Object.entries(original.header ?? {})
      .filter(([_, value]) => value !== null)
      .map(([key, value]) => value?.map(v => `${key}: ${v}`).join('\n') ?? '')
      .join('\n'),
  )

  // These are the form-binded values
  // svelte-ignore state_referenced_locally
  let query = $state<string>(originalQuery)
  // svelte-ignore state_referenced_locally
  let header = $state<string>(originalHeader)

  let cronScheduler: CronScheduler | undefined = $state(undefined)

  // This is called by Instances.svelte when the reset button is clicked.
  export const reset = () => {
    cronScheduler?.reset()
    header = Object.entries(form.header ?? {})
      .filter(([_, value]) => value !== null)
      .map(([key, value]) => value?.map(v => `${key}: ${v}`).join('\n') ?? '')
      .join('\n')

    query = Object.entries(form.query ?? {})
      .filter(([_, value]) => value !== null)
      .map(([key, value]) => value?.map(v => `${key}=${v}`).join('\n') ?? '')
      .join('\n')
  }

  const updateMap = (data: string, split: string): Record<string, string[]> => {
    const reduce = (acc: Record<string, string[]>, line: string) => {
      const [key, value] = line.split(split).map(s => s.trim())
      if (key && value) acc[key] = [...(acc[key] ?? []), value]
      return acc
    }

    return data
      .split('\n')
      .filter(l => l.trim())
      .reduce(reduce, {})
  }
</script>

<div class="endpoint">
  <Row>
    <Col md={12}>
      <Input
        id={app.id + '.name'}
        bind:value={form.name}
        original={original?.name}
        {validate} />
    </Col>
    <Col md={6}>
      <Input
        id={app.id + '.template'}
        bind:value={form.template}
        original={original?.template}
        {validate}>
        {#snippet post()}
          <Button
            color="secondary"
            outline
            style="width:44px;"
            class={form.follow === original.follow ? '' : 'changed'}>
            <Box type="checkbox" id={app.id + '.follow'} bind:checked={form.follow} />
          </Button>
        {/snippet}
      </Input>
    </Col>
    <Col md={6}>
      <Input
        type="timeout"
        id={app.id + '.timeout'}
        bind:value={form.timeout}
        original={original?.timeout}
        noDisable
        {validate} />
    </Col>
    <Col md={2}>
      <Input
        type="select"
        id={app.id + '.method'}
        bind:value={form.method}
        original={original?.method}
        {validate}
        options={['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'].map(m => ({
          name: m,
          value: m,
        }))} />
    </Col>
    <Col md={10}>
      <CheckedInput id="url" {app} {index} bind:form bind:original {validate} />
    </Col>
    <Col lg={12}>
      <Input
        rows={Math.min(form.body?.split('\n').length ?? 1, 15)}
        type="textarea"
        id={app.id + '.body'}
        bind:value={form.body}
        original={original?.body}
        badge={$_('scheduler.badge.body', { values: { count: form.body?.length ?? 0 } })}
        {validate} />
    </Col>
    <Col lg={12}>
      <Input
        rows={Math.min(query.split('\n').length, 15)}
        type="textarea"
        id={app.id + '.query'}
        bind:value={query}
        original={originalQuery}
        badge={$_('scheduler.badge.query', {
          values: { count: mapLength(form.query ?? {}) },
        })}
        validate={(id, value) => {
          form.query = updateMap(query, '=')
          return validate?.(id, value) ?? ''
        }} />
    </Col>
    <Col lg={12}>
      <Input
        rows={Math.min(header.split('\n').length, 15)}
        type="textarea"
        id={app.id + '.header'}
        bind:value={header}
        original={originalHeader}
        badge={$_('scheduler.badge.header', {
          values: { count: mapLength(form.header ?? {}) },
        })}
        validate={(id, value) => {
          form.header = updateMap(header, ': ')
          return validate?.(id, value) ?? ''
        }} />
    </Col>

    <Col md={12}>
      <CronScheduler bind:cron={form} {original} {validate} bind:this={cronScheduler} />
    </Col>
  </Row>
</div>

<style>
  .endpoint :global(.changed) {
    background-color: rgba(205, 92, 92, 0.322) !important;
  }
</style>
