<!-- This input component is used for URLs and MySQL hostnames in forms. -->
<script lang="ts">
  import { Button, Alert, Input as Box } from '@sveltestrap/sveltestrap'
  import { _ } from './Translate.svelte'
  import {
    faCircleCheck,
    faCircleXmark,
    faSpinner,
  } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { faCheckDouble } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { get } from 'svelte/store'
  import type { App } from './formsTracker.svelte'
  import Input from './Input.svelte'
  import { delay, maxLength } from './util'
  import { postUi } from '../api/fetch'
  import Fa from './Fa.svelte'
  import { slide } from 'svelte/transition'

  type Props<T> = {
    id: keyof T
    app: App<T>
    index: number
    form: T
    original: T
    /** Disable the check button.*/
    disabled?: boolean
    params?: () => Promise<URLSearchParams>
    [key: string]: any
  }
  let {
    id,
    app,
    index,
    form = $bindable(),
    original = $bindable(),
    disabled = false,
    params = undefined,
    ...rest
  }: Props<any> = $props()

  // Used for instance checking.
  let ok = $state(undefined as boolean | undefined)
  let body = $state('')
  let testing = $state(false)

  const checkInstance = async (e: Event) => {
    e.preventDefault()
    body = ''
    testing = true

    let p = ''
    try {
      if (params) p = `?${(await params()).toString()}`
      else await delay(300) // satisfying spinner.
    } catch {
      testing = false
      return
    }

    const uri = 'checkInstance/' + app.name.toLowerCase() + '/' + index + p
    const data = app.merge(index, form)
    const res = await postUi(uri, JSON.stringify(data), false)
    ok = res.ok
    body = res.body
    testing = false
  }
</script>

{#snippet validSsl()}
  <Button
    style="width:44px;"
    type="button"
    outline
    color="notifiarr"
    class={form['validSsl'] !== original['validSsl'] ? 'changed' : ''}
    onclick={() => (form['validSsl'] = !form['validSsl'])}>
    <Box type="checkbox" bind:checked={form['validSsl']} />
  </Button>
{/snippet}

{#snippet shell()}
  <Button
    style="width:44px;"
    type="button"
    outline
    color="notifiarr"
    class={form['shell'] !== original['shell'] ? 'changed' : ''}
    onclick={() => (form['shell'] = !form['shell'])}>
    <Box type="checkbox" bind:checked={form['shell']} />
  </Button>
{/snippet}

<div class="checked-input">
  <Input
    id={app.id + '.' + id.toString()}
    bind:value={form[id]}
    original={original?.[id] ?? undefined}
    disabled={app.disabled?.includes(id.toString())}
    description={id === 'url' && form[id]?.toString()?.startsWith('https://')
      ? get(_)('words.instance-options.validSsl.description')
      : rest.description}
    {...rest}>
    <!-- This is a "checked" input, so add a check button for the instance. -->
    {#snippet pre()}
      <Button
        style="width:44px;"
        type="button"
        outline
        color="notifiarr"
        disabled={testing || disabled}
        onclick={checkInstance}>
        {#if testing}
          <Fa i={faSpinner} c1="orange" spin scale={1.5} />
        {:else}
          <Fa
            i={faCheckDouble}
            c1={disabled ? 'lightgrey' : 'green'}
            c2={disabled ? 'darkgrey' : 'darkcyan'}
            d1={disabled ? 'darkgrey' : 'limegreen'}
            d2={disabled ? 'lightgrey' : 'cyan'}
            scale={1.5} />
        {/if}
      </Button>
    {/snippet}

    <!-- If they type in an https:// url, add a checkbox to validate the SSL certificate. -->
    {#snippet post()}
      {#if id === 'url' && form[id]?.startsWith('https://')}
        {@render validSsl()}
      {:else if id === 'command'}
        {@render shell()}
      {/if}
    {/snippet}

    <!-- Feedback message is only used when the test/check button is clicked. -->
    {#snippet msg()}
      {#if body}
        <div transition:slide>
          <Alert
            fade={false}
            isOpen
            toggle={() => (body = '')}
            color={ok ? 'success' : 'danger'}>
            <Fa
              scale={1.5}
              i={ok ? faCircleCheck : faCircleXmark}
              c1={ok ? 'green' : 'firebrick'}
              c2="white"
              d2="black" /> &nbsp; {@html maxLength(body, 200)}
          </Alert>
        </div>
      {/if}
    {/snippet}
  </Input>
</div>

<style>
  .checked-input :global(.changed) {
    background-color: rgba(205, 92, 92, 0.322);
  }
</style>
