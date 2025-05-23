<script lang="ts">
  import { get } from 'svelte/store'
  import Input from './Input.svelte'
  import CheckedInput from './CheckedInput.svelte'
  import { Col, Row, Button } from '@sveltestrap/sveltestrap'
  import { _ } from './Translate.svelte'
  import { deepCopy, deepEqual } from './util'
  import type { Config } from '../api/notifiarrConfig'
  import { slide } from 'svelte/transition'

  export type App = {
    id: string
    name: string
    logo: string
    disabled?: string[]
    hidden?: string[]
    empty?: any
    merge: (index: number, form: Form) => Config
  }

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

  type Props = {
    form: Form
    original: Form
    app: App
    index?: number
    resetButton?: boolean
    validate?: (id: string, index: number, value: any, reset?: boolean) => string
  }

  let {
    form = $bindable(),
    original,
    app,
    index = 0,
    resetButton = true,
    validate,
  }: Props = $props()

  // Convert array to newline-separated string for textarea.
  let busIds = $state(
    typeof form.busIDs === 'undefined' ? undefined : form.busIDs?.join('\n'),
  )

  const feedback = $state<Record<string, string>>({})
  const rows = $derived(
    busIds ? (busIds.split('\n').length < 10 ? busIds?.split('\n').length : 1) : 1,
  )

  // Used a shorthand variable to set Col sizes.
  const hasToken = $derived(
    typeof form.token === 'string' || typeof form.apiKey === 'string',
  )
  const validateApp = (id: string, val: any) => validate?.(id, index, val) ?? ''
  export const valid = () => Object.values(feedback).every(v => !v)

  export const resetFeedback = () => {
    // Calling validate with true at the end will delete all the feedback.
    validate?.(app.id + '.all', index, 'reset', true)
  }

  // Reset the form and feedback.
  export const reset = (e?: Event, deleted?: boolean) => {
    e?.preventDefault()
    // Copy form, and reset all the validators.
    if (!deleted) form = deepCopy(original)
    Object.keys(form).forEach(id => {
      feedback[app.id + '.' + id] = deleted
        ? ''
        : (validate?.(app.id + '.' + id, index, form[id as keyof Form]) ?? '')
    })
  }

  $effect(() => {
    // Only update form variables if they existed prior to this effect.
    if (typeof form.disabled !== 'undefined') form.busIDs = busIds?.split(/\s+/)
  })
</script>

<div class="instance">
  <!-- Top row, shows name and url or hostname/ip. -->
  <Row>
    {#if typeof form.name === 'string'}
      <Col lg={6} xl={hasToken ? 4 : 6}>
        <Input
          id={app.id + '.name'}
          bind:value={form.name}
          bind:feedback={feedback[app.id + '.name']}
          original={original?.name}
          disabled={app.disabled?.includes('name')}
          validate={validateApp} />
      </Col>
    {/if}
    {#if typeof form.url === 'string' && !app.hidden?.includes('url')}
      <Col lg={6} xl={hasToken ? 4 : 6}>
        <CheckedInput
          id="url"
          bind:feedback={feedback[app.id + '.url']}
          bind:original
          disabled={app.disabled?.includes('url')}
          {app}
          {form}
          {index}
          validate={validateApp} />
      </Col>
    {/if}

    {#if typeof form.host === 'string' && !app.hidden?.includes('host')}
      <Col lg={6} xl={hasToken ? 4 : 6}>
        <CheckedInput
          id="host"
          bind:feedback={feedback[app.id + '.host']}
          bind:original
          disabled={app.disabled?.includes('host')}
          {app}
          {form}
          {index}
          validate={validateApp} />
      </Col>
    {/if}
    {#if typeof form.apiKey === 'string'}
      <Col lg={12} xl={4}>
        <Input
          id={app.id + '.apiKey'}
          type="password"
          bind:value={form.apiKey}
          bind:feedback={feedback[app.id + '.apiKey']}
          original={original?.apiKey}
          disabled={app.disabled?.includes('apiKey')}
          validate={validateApp} />
      </Col>
    {/if}
    {#if typeof form.token === 'string' && !app.hidden?.includes('token')}
      <Col lg={12} xl={4}>
        <Input
          id={app.id + '.token'}
          type="password"
          bind:value={form.token}
          bind:feedback={feedback[app.id + '.token']}
          original={original?.token}
          disabled={app.disabled?.includes('token')}
          validate={validateApp} />
      </Col>
    {/if}
  </Row>

  <Row>
    {#if typeof form.username === 'string'}
      <Col lg={6} xl={6}>
        <Input
          id={app.id + '.username'}
          bind:value={form.username}
          bind:feedback={feedback[app.id + '.username']}
          original={original?.username}
          disabled={app.disabled?.includes('username')}
          validate={validateApp} />
      </Col>
    {/if}
    {#if typeof form.password === 'string' && !app.hidden?.includes('password')}
      <Col lg={6} xl={6}>
        <Input
          id={app.id + '.password'}
          type="password"
          bind:value={form.password}
          bind:feedback={feedback[app.id + '.password']}
          original={original?.password}
          disabled={app.disabled?.includes('password')}
          validate={validateApp} />
      </Col>
    {/if}
  </Row>

  <Row>
    {#if typeof form.timeout === 'string'}
      <Col md={!app.hidden?.includes('deletes') ? 4 : 6}>
        <Input
          id="words.instance-options.timeout"
          type="timeout"
          bind:value={form.timeout}
          bind:feedback={feedback[app.id + '.timeout']}
          original={original?.timeout}
          disabled={app.disabled?.includes('timeout')}
          validate={validateApp} />
      </Col>
    {/if}
    {#if typeof form.interval === 'string'}
      <Col md={!app.hidden?.includes('deletes') ? 4 : 6}>
        <Input
          id="words.instance-options.interval"
          type="interval"
          bind:value={form.interval}
          bind:feedback={feedback[app.id + '.interval']}
          original={original?.interval}
          disabled={app.disabled?.includes('interval')}
          validate={validateApp} />
      </Col>
    {/if}
    {#if !app.hidden?.includes('deletes')}
      <Col md={4}>
        <Input
          id="words.instance-options.deletes"
          type="select"
          bind:value={form.deletes}
          bind:feedback={feedback[app.id + '.deletes']}
          original={original?.deletes}
          disabled={app.disabled?.includes('deletes')}
          validate={validateApp}>
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

  <Row>
    {#if typeof form.disabled === 'boolean'}
      <Col lg={4}>
        <Input
          id={app.id + '.disabled'}
          type="select"
          bind:value={form.disabled}
          bind:feedback={feedback[app.id + '.disabled']}
          original={original?.disabled}
          disabled={app.disabled?.includes('disabled')}
          validate={validateApp}>
          <!-- These are backward on purpose.-->
          <option value={false} selected={form.disabled === false}>
            {$_('words.select-option.Enabled')}
          </option>
          <option value={true} selected={form.disabled === true}>
            {$_('words.select-option.Disabled')}
          </option>
        </Input>
      </Col>
    {/if}
    {#if typeof form.smiPath === 'string'}
      <Col lg={4}>
        <CheckedInput
          id="smiPath"
          bind:feedback={feedback[app.id + '.smiPath']}
          bind:original
          disabled={app.disabled?.includes('smiPath')}
          {app}
          {form}
          {index}
          {validate} />
      </Col>
      <Col lg={4}>
        <Input
          id={app.id + '.busIds'}
          type="textarea"
          {rows}
          bind:value={busIds}
          bind:feedback={feedback[app.id + '.busIds']}
          original={original?.busIDs?.join('\n')}
          disabled={app.disabled?.includes('busIds')}
          validate={validateApp} />
      </Col>
    {/if}
  </Row>

  {#if resetButton && !deepEqual(form, original)}
    <div class="mb-2" transition:slide>
      <Button color="primary" outline onclick={reset} class="float-end">
        {$_('buttons.ResetForm')}
      </Button>
      &nbsp;
    </div>
  {/if}
</div>
