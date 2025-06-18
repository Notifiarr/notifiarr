<!-- This handles all the standard app types. Does not handle non-app things like file watcher and endpoints. -->
<script lang="ts" module>
  // This is a combination of all types supported: starr, downloaders, media, snapshot, etc.
  export type Form = {
    name?: string
    url?: string
    host?: string
    username?: string
    password?: string
    apiKey?: string
    token?: string
    timeout?: string
    interval?: string
    deletes?: number
    validSsl?: boolean
    /* Nvidia only */
    smiPath?: string
    busIDs?: string[]
    disabled?: boolean
  }
</script>

<script lang="ts">
  import { get } from 'svelte/store'
  import Input from './Input.svelte'
  import CheckedInput from './CheckedInput.svelte'
  import { Col, Row, Button } from '@sveltestrap/sveltestrap'
  import { _ } from './Translate.svelte'
  import { deepEqual } from './util'
  import { slide } from 'svelte/transition'
  import type { ChildProps } from './Instances.svelte'

  let {
    form = $bindable(),
    original,
    app,
    index = 0,
    validate,
    reset,
  }: ChildProps<any> = $props()

  /** Array converted to newline-separated string for textarea. */
  let busIds: string = $state(form.busIDs?.join('\n') ?? '')
  /** Rows for textarea based on count of newlines. */
  const rows = $derived(busIds.split('\n').length < 10 ? busIds.split('\n').length : 1)
  // Update form variables if they existed prior to this effect.
  $effect(() => {
    // form.smiPath and form.busIDs only exist on Nvidia.
    if (typeof form.smiPath !== 'undefined') form.busIDs = busIds.split(/\s+/)
  })

  /** Shorthand variable to set Col sizes for username and password inputs. */
  const hasToken = $derived(
    (typeof form.token === 'string' || typeof form.apiKey === 'string') &&
      !app.hidden?.includes('token') &&
      !app.hidden?.includes('apiKey'),
  )
</script>

<div class="instance">
  <!-- Top row, shows name and url or hostname/ip. -->
  <Row>
    <!-- Name is required for all integrations except Nvidia. -->
    {#if typeof form.name === 'string'}
      <Col lg={6} xl={hasToken ? 4 : 6}>
        <Input
          id={app.id + '.name'}
          bind:value={form.name}
          original={original?.name}
          disabled={app.disabled?.includes('name')}
          {validate} />
      </Col>
    {/if}
    {#if typeof form.url === 'string' && !app.hidden?.includes('url')}
      <Col lg={6} xl={hasToken ? 4 : 6}>
        <CheckedInput
          id="url"
          bind:original
          bind:form
          disabled={app.disabled?.includes('url')}
          {app}
          {index}
          {validate} />
      </Col>
    {/if}

    <!-- Host is only used by MySQL, and it's treated similar to a URL. -->
    {#if typeof form.host === 'string' && !app.hidden?.includes('host')}
      <Col lg={6} xl={hasToken ? 4 : 6}>
        <CheckedInput
          id="host"
          bind:original
          bind:form
          disabled={app.disabled?.includes('host')}
          {app}
          {index}
          {validate} />
      </Col>
    {/if}
    {#if typeof form.apiKey === 'string' && !app.hidden?.includes('apiKey')}
      <Col lg={12} xl={4}>
        <Input
          id={app.id + '.apiKey'}
          type="password"
          bind:value={form.apiKey}
          original={original?.apiKey}
          disabled={app.disabled?.includes('apiKey')}
          {validate} />
      </Col>
    {/if}
    <!-- Plex uses a token, and not an api key. It's one or the other in the form (or neither in a few cases). -->
    {#if typeof form.token === 'string' && !app.hidden?.includes('token')}
      <Col lg={12} xl={4}>
        <Input
          id={app.id + '.token'}
          type="password"
          bind:value={form.token}
          original={original?.token}
          disabled={app.disabled?.includes('token')}
          {validate} />
      </Col>
    {/if}
  </Row>

  <Row>
    <!-- In a rare case (deluge) there is no username, so make the password input wider. -->
    {#if typeof form.username === 'string' && !app.hidden?.includes('username')}
      <Col md={app.hidden?.includes('password') ? 12 : 6}>
        <Input
          id={app.id + '.username'}
          bind:value={form.username}
          original={original?.username}
          disabled={app.disabled?.includes('username')}
          {validate} />
      </Col>
    {/if}
    <!-- The process is repeated for the username, but no case exists where there is no password but there is a username. -->
    {#if typeof form.password === 'string' && !app.hidden?.includes('password')}
      <Col md={app.hidden?.includes('username') ? 12 : 6}>
        <Input
          id={app.id + '.password'}
          type="password"
          bind:value={form.password}
          original={original?.password}
          disabled={app.disabled?.includes('password')}
          {validate} />
      </Col>
    {/if}
  </Row>

  <Row>
    <!-- If there's no delete, then timeout and interval are wider.-->
    {#if typeof form.timeout === 'string'}
      <Col md={!app.hidden?.includes('deletes') ? 4 : 6}>
        <Input
          id={app.id + '.timeout'}
          description={$_('words.instance-options.timeout.description')}
          label={$_('words.instance-options.timeout.label')}
          type="timeout"
          bind:value={form.timeout}
          original={original?.timeout}
          disabled={app.disabled?.includes('timeout')}
          {validate} />
      </Col>
    {/if}
    {#if typeof form.interval === 'string'}
      <Col md={!app.hidden?.includes('deletes') ? 4 : 6}>
        <Input
          id={app.id + '.interval'}
          description={$_('words.instance-options.interval.description')}
          label={$_('words.instance-options.interval.label')}
          type="interval"
          bind:value={form.interval}
          original={original?.interval}
          disabled={app.disabled?.includes('interval')}
          {validate} />
      </Col>
    {/if}

    <!-- Starr only -->
    {#if !app.hidden?.includes('deletes')}
      <Col md={4}>
        <Input
          id={app.id + '.deletes'}
          description={$_('words.instance-options.deletes.description')}
          tooltip={$_('words.instance-options.deletes.tooltip')}
          label={$_('words.instance-options.deletes.label')}
          type="select"
          bind:value={form.deletes}
          original={original?.deletes}
          disabled={app.disabled?.includes('deletes')}
          {validate}>
          <option value={0}>{get(_)('words.select-option.Disabled')}</option>
          {#each ['1', '2', '5', '7', '10', '15', '20', '50', '100', '200'] as count}
            <option value={count}>
              {get(_)('words.instance-options.deletes.countPerHour', {
                values: { count },
              })}
            </option>
          {/each}
        </Input>
      </Col>
    {/if}
  </Row>

  <!-- Nvidia only-->
  <Row>
    {#if typeof form.disabled === 'boolean'}
      <Col lg={4}>
        <Input
          id={app.id + '.disabled'}
          type="select"
          bind:value={form.disabled}
          original={original?.disabled}
          disabled={app.disabled?.includes('disabled')}
          {validate} />
      </Col>
    {/if}
    {#if typeof form.smiPath === 'string'}
      <Col lg={4}>
        <CheckedInput
          id="smiPath"
          bind:original
          bind:form
          disabled={app.disabled?.includes('smiPath')}
          {app}
          {index}
          {validate} />
      </Col>
      <Col lg={4}>
        <Input
          id={app.id + '.busIds'}
          type="textarea"
          {rows}
          bind:value={busIds}
          original={original?.busIDs?.join('\n') ?? ''}
          disabled={app.disabled?.includes('busIds')}
          {validate} />
      </Col>
    {/if}
  </Row>

  <!-- Show an optional reset button if the form has changes. -->
  {#if reset && !deepEqual(form, original)}
    <div class="mb-2" transition:slide>
      <Button color="primary" outline onclick={reset} class="float-end">
        {$_('buttons.ResetForm')}
      </Button>
      &nbsp;
    </div>
  {/if}
</div>
