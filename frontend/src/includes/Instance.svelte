<script lang="ts">
  import { get } from 'svelte/store'
  import Input from './Input.svelte'
  import { Col, Row, Input as Box, Button, Alert } from '@sveltestrap/sveltestrap'
  import { _ } from './Translate.svelte'
  import { postUi } from '../api/fetch'
  import {
    faCircleCheck,
    faCircleXmark,
    faSpinner,
  } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { faCheckDouble } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from './Fa.svelte'
  import { slide } from 'svelte/transition'
  import { delay } from './util'
  import type { Config } from '../api/notifiarrConfig'

  export type App = {
    id: string
    name: string
    logo: string
    disabled?: string[]
    hidden?: string[]
  }

  // This is a combination of all types supported: starr, downloaders, media, snapshot, etc.
  export type Form = {
    name: string
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
  }

  type Props = { form: Form; app: App; index: number }
  let { form = $bindable(), app, index }: Props = $props()
  let ok = $state(undefined as boolean | undefined)
  let body = $state('')
  let testing = $state(false)

  const checkInstance = async (e: Event) => {
    e.preventDefault()
    body = ''
    testing = true
    await delay(300) // satisfying spinner.
    const uri = 'checkInstance/' + app.name.toLowerCase() + '/' + index
    const data = { ...({} as Config), [app.name.toLowerCase()]: form }
    const res = await postUi(uri, JSON.stringify(data), false)
    ok = res.ok
    body = res.body
    testing = false
  }
</script>

<!-- Top row, shows name and url or hostname/ip. -->
<Row>
  <Col lg={6} xl={4}>
    <Input
      id={app.id + '.name'}
      type="text"
      bind:value={form.name}
      disabled={app.disabled?.includes('name')} />
  </Col>
  {#if typeof form.url === 'string' && !app.hidden?.includes('url')}
    <Col lg={6} xl={4}>
      <Input
        id={app.id + '.url'}
        type="text"
        bind:value={form.url}
        description={form.url?.startsWith('https://')
          ? get(_)('words.instance-options.validSsl.description')
          : undefined}
        disabled={app.disabled?.includes('url')}>
        {#snippet pre()}
          <Button
            type="button"
            outline
            color="notifiarr"
            onclick={checkInstance}
            disabled={testing}>
            {#if testing}
              <Fa i={faSpinner} c1="orange" spin scale={1.5} />
            {:else}
              <Fa i={faCheckDouble} c1="green" c2="cyan" scale={1.5} />
            {/if}
          </Button>
        {/snippet}
        <!-- If they type in an https:// url, add a checkbox to validate the SSL certificate. -->
        {#snippet post()}
          {#if form.url?.startsWith('https://')}
            <Button
              type="button"
              outline
              color="notifiarr"
              onclick={() => (form.validSsl = !form.validSsl)}>
              <Box type="checkbox" bind:checked={form.validSsl} />
            </Button>
          {/if}
        {/snippet}
        {#snippet msg()}
          {#if body}
            <div transition:slide>
              <Alert
                fade={false}
                isOpen={!!body}
                toggle={() => (body = '')}
                color={ok ? 'success' : 'danger'}>
                <Fa
                  scale={1.5}
                  i={ok ? faCircleCheck : faCircleXmark}
                  c1={ok ? 'green' : 'firebrick'}
                  c2="white"
                  d2="black" /> &nbsp; {body}
              </Alert>
            </div>
          {/if}
        {/snippet}
      </Input>
    </Col>
  {/if}
  {#if typeof form.host === 'string' && !app.hidden?.includes('host')}
    <Col lg={6} xl={4}>
      <Input
        id={app.id + '.host'}
        type="text"
        bind:value={form.host}
        disabled={app.disabled?.includes('host')} />
    </Col>
  {/if}
  {#if typeof form.apiKey === 'string'}
    <Col lg={12} xl={4}>
      <Input
        id={app.id + '.apiKey'}
        type="password"
        bind:value={form.apiKey}
        disabled={app.disabled?.includes('apiKey')} />
    </Col>
  {/if}
  {#if typeof form.token === 'string' && !app.hidden?.includes('token')}
    <Col lg={12} xl={4}>
      <Input
        id={app.id + '.token'}
        type="password"
        bind:value={form.token}
        disabled={app.disabled?.includes('token')} />
    </Col>
  {/if}
</Row>

<Row>
  {#if typeof form.username === 'string'}
    <Col lg={6} xl={4}>
      <Input
        id={app.id + '.username'}
        type="text"
        bind:value={form.username}
        disabled={app.disabled?.includes('username')} />
    </Col>
  {/if}
  {#if typeof form.password === 'string' && !app.hidden?.includes('password')}
    <Col lg={6} xl={4}>
      <Input
        id={app.id + '.password'}
        type="password"
        bind:value={form.password}
        disabled={app.disabled?.includes('password')} />
    </Col>
  {/if}
</Row>

<Row>
  <Col md={!app.hidden?.includes('deletes') ? 4 : 6}>
    <Input id="words.instance-options.timeout" type="timeout" bind:value={form.timeout} />
  </Col>
  <Col md={!app.hidden?.includes('deletes') ? 4 : 6}>
    <Input
      id="words.instance-options.interval"
      type="interval"
      bind:value={form.interval} />
  </Col>
  {#if !app.hidden?.includes('deletes')}
    <Col md={4}>
      <Input
        id="words.instance-options.deletes"
        type="select"
        bind:value={form.deletes}
        disabled={app.disabled?.includes('deletes')}>
        <option value={0}>{get(_)('words.select-option.Disabled')}</option>
        {#each ['1', '2', '5', '7', '10', '15', '20', '50', '100', '200'] as count}
          <option value={count}>
            {get(_)('words.instance-options.deletes.countPerHour', { values: { count } })}
          </option>
        {/each}
      </Input>
    </Col>
  {/if}
</Row>
