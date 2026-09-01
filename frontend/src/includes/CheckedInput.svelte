<!-- This input component is used for URLs and MySQL hostnames in forms. -->
<script lang="ts">
  import { Button, Alert, Input as Box } from '@sveltestrap/sveltestrap'
  import { _ } from './Translate.svelte'
  import CircleNotch from 'phosphor-svelte/lib/CircleNotch'
  import Checks from 'phosphor-svelte/lib/Checks'
  import CheckCircle from 'phosphor-svelte/lib/CheckCircle'
  import XCircle from 'phosphor-svelte/lib/XCircle'
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
    envVar: string
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
    envVar,
    ...rest
  }: Props<any> = $props()

  // Used for instance checking.
  let ok = $state(undefined as boolean | undefined)
  let body = $state('')
  let testing = $state(false)
  const https = $derived(id === 'url' && form?.[id]?.toString()?.startsWith('https://'))

  const checkInstance = async (e: Event) => {
    e.preventDefault()
    if (!form) return
    body = ''
    testing = true
    try {
      let p = ''
      if (params) p = `?${(await params()).toString()}`
      else await delay(300) // satisfying spinner.
      const uri = 'checkInstance/' + app.name.toLowerCase() + '/' + index + p
      const data = app.merge(index, form)
      const res = await postUi(uri, JSON.stringify(data), false)
      ok = res.ok
      body = res.body
    } catch {
      // params() can throw when extra query data is not ready.
    } finally {
      testing = false
    }
  }
</script>

{#snippet validSsl()}
  <Button
    style="width:44px;"
    type="button"
    outline
    title={$_('phrases.ValidateSSL')}
    aria-pressed={!!form?.['validSsl']}
    color="notifiarr"
    class={form?.['validSsl'] !== original?.['validSsl'] ? 'changed' : ''}
    onclick={() => form && (form['validSsl'] = !form['validSsl'])}>
    <Box
      type="checkbox"
      checked={!!form['validSsl']}
      tabindex={-1}
      style="pointer-events:none" />
  </Button>
{/snippet}

{#snippet shell()}
  <Button
    style="width:44px;"
    type="button"
    outline
    title={$_('phrases.RunAsShell')}
    aria-pressed={!!form?.['shell']}
    color="notifiarr"
    class={form?.['shell'] !== original?.['shell'] ? 'changed' : ''}
    onclick={() => form && (form['shell'] = !form['shell'])}>
    <Box
      type="checkbox"
      checked={!!form['shell']}
      tabindex={-1}
      style="pointer-events:none" />
  </Button>
{/snippet}

{#if form}
  <div class="checked-input">
    <Input
      id={app.id + '.' + id.toString()}
      bind:value={form[id]}
      original={original?.[id] ?? undefined}
      disabled={app.disabled?.includes(id.toString())}
      description={https
        ? $_('words.instance-options.validSsl.description')
        : rest.description}
      {envVar}
      {...rest}>
      <!-- This is a "checked" input, so add a check button for the instance. -->
      {#snippet pre()}
        <Button
          style="width:44px;"
          type="button"
          outline
          title={$_('phrases.CheckInstance')}
          color="notifiarr"
          disabled={testing || disabled}
          onclick={checkInstance}>
          {#if testing}
            <Fa i={CircleNotch} c1="orange" spin btn />
          {:else}
            <Fa
              i={Checks}
              c1={disabled ? 'lightgrey' : 'green'}
              d1={disabled ? 'darkgrey' : 'limegreen'}
              btn />
          {/if}
        </Button>
      {/snippet}

      <!-- If they type in an https:// url, add a checkbox to validate the SSL certificate. -->
      {#snippet post()}
        {#if https}
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
                btn
                i={ok ? CheckCircle : XCircle}
                c1={ok ? 'green' : 'firebrick'}
                c2="white"
                d2="black">{@html maxLength(body, 200)}</Fa>
            </Alert>
          </div>
        {/if}
      {/snippet}
    </Input>
  </div>
{/if}
