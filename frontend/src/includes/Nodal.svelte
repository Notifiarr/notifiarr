<!--
 Wrap the sveltestrap Modal component to add a class name when the modal is open.
 Works around a bug. Also gives us a place to add business logic.

 Business Logic:
 - Title expansion from translation key.
 - Title buttons: Refresh, Raw, Fullscreen, Close.
 - Ability to get and refresh the backend response.
 -->

<script lang="ts">
  import { type BackendResponse } from '../api/fetch'
  import {
    Button,
    ButtonGroup,
    ModalBody,
    ModalFooter,
    ModalHeader,
    Popover,
    Modal,
  } from '@sveltestrap/sveltestrap'
  import { theme } from './theme.svelte'
  import T, { _ } from './Translate.svelte'
  import Fa from './Fa.svelte'
  import type { Props as FaProps } from './Fa.svelte'
  import { tick, type Snippet } from 'svelte'
  import ArrowsClockwise from 'phosphor-svelte/lib/ArrowsClockwise'
  import ArrowsIn from 'phosphor-svelte/lib/ArrowsIn'
  import ArrowsOut from 'phosphor-svelte/lib/ArrowsOut'
  import ArrowsLeftRight from 'phosphor-svelte/lib/ArrowsLeftRight'
  import X from 'phosphor-svelte/lib/X'
  import { delay } from './util'

  type Props = {
    /** Children is rendered in a modal body. Backend response is passed as the first argument. */
    children: Snippet<[BackendResponse | undefined]>
    /** Icon to display in the modal header. */
    fa?: FaProps
    /** Title to display in the modal header. This is a full translation key. */
    title?: string
    /** Values to pass to the title translation. */
    values?: Record<string, any>
    /** Whether the modal is closable by pressing the escape key. */
    esc?: boolean
    /** Optional function to get the backend response. */
    get?: () => Promise<BackendResponse>
    /** Footer to render in the modal footer. */
    footer?: Snippet<[BackendResponse | undefined]>
    /** Whether the modal is open. */
    isOpen?: boolean
    /** Whether the modal is full screen height.  */
    full?: boolean
    /** Whether the modal takes up the whole screen. Bindable because the user can toggle it. */
    fullscreen?: boolean
    /** Optional function to call when the modal is closed. */
    follow?: () => void
    /** Whether the modal is closable. */
    disabled?: boolean
    [key: string]: any
  }

  let {
    children,
    fa,
    title = '',
    values,
    get,
    footer,
    isOpen = $bindable(false),
    full = false,
    fullscreen = $bindable(false),
    esc = false,
    follow,
    disabled = false,
    ...rest
  }: Props = $props()

  let loading = $state(false)
  let resp = $state<BackendResponse>()
  let showRaw = $state(false)
  let ok = $state(false) // we use ok to keep the refresh from running away

  const height = $derived('calc(100vh - ' + (footer ? '195' : '125') + 'px)')
  const modalClassName = $derived(rest.modalClassName + (isOpen ? ' show' : ''))

  export const close = async (e?: Event) => {
    e?.preventDefault()
    await follow?.()
    await tick()
    if (isOpen) isOpen = false
  }

  export const open = (e?: Event) => {
    e?.preventDefault()
    isOpen = true
  }

  const refresh = async () => {
    try {
      loading = true
      resp = get ? (await delay(300), await get()) : undefined
    } catch (error) {
      resp = { ok: false, body: error }
    } finally {
      loading = false
      if (ok != resp?.ok) ok = resp?.ok ?? false
    }
  }

  $effect(() => {
    if (isOpen && !ok) refresh()
  })
</script>

{#snippet heading()}
  {#if title}
    {@const t = $_(title, { values: values ?? undefined })}
    {#if typeof t !== 'string'}{t['title']}{:else}{t}{/if}
  {/if}
{/snippet}

<Modal
  {...{ ...rest, isOpen, modalClassName, theme: $theme, fullscreen }}
  toggle={esc && !disabled ? close : undefined}>
  <ModalHeader class="d-inline-block">
    <!-- Header title and icon. -->
    {#if fa && title}
      <Fa {...fa} logo>{@render heading()}</Fa>
    {:else if fa}
      <Fa {...fa} logo />
    {:else}
      {@render heading()}
    {/if}
    <!-- Header buttons. -->
    {#if isOpen}
      <ButtonGroup class="float-end">
        {#if get}
          <Button
            type="button"
            id="refreshM"
            outline
            color="secondary"
            size="sm"
            onclick={refresh}
            aria-label={$_('Nodal.button.refresh')}
            title={$_('Nodal.button.refresh')}>
            <Fa
              i={ArrowsClockwise}
              c1="steelblue"
              c2="firebrick"
              d2="pink"
              btn
              spin={loading} />
          </Button>
          <Popover target="refreshM" trigger="hover" theme={$theme}>
            <T id="Nodal.button.refresh" />
          </Popover>
          {#if resp && resp.ok}
            <Button
              type="button"
              id="rawM"
              outline
              color="secondary"
              size="sm"
              onclick={() => (showRaw = !showRaw)}>
              <Fa i={ArrowsLeftRight} c1="steelblue" c2="firebrick" d2="pink" btn />
            </Button>
            <Popover target="rawM" trigger="hover" theme={$theme}>
              <T id="Nodal.button.raw" />
            </Popover>
          {/if}
        {/if}

        <Button
          type="button"
          id="fullscreenM"
          outline
          color="secondary"
          size="sm"
          onclick={() => (fullscreen = !fullscreen)}
          aria-label={$_('Nodal.button.fullscreen')}
          title={$_('Nodal.button.fullscreen')}>
          <Fa
            i={fullscreen ? ArrowsIn : ArrowsOut}
            c1="steelblue"
            c2="firebrick"
            d2="pink"
            btn />
        </Button>
        <Popover target="fullscreenM" trigger="hover" theme={$theme}>
          <T id="Nodal.button.fullscreen" />
        </Popover>

        <Button
          type="button"
          id="closeM"
          outline
          color="secondary"
          size="sm"
          title={$_('buttons.Close')}
          aria-label={$_('buttons.Close')}
          onclick={close}
          {disabled}>
          <Fa i={X} c2="orange" d2="gold" btn />
        </Button>
      </ButtonGroup>
    {/if}
  </ModalHeader>

  <form onsubmit={e => e.preventDefault()}>
    <ModalBody style="max-height: {height}; overflow: auto;">
      {#if loading && !resp}
        <T id="phrases.Loading" />
      {:else if resp && resp.ok && showRaw}
        <pre class="pre-wrap" style="overflow: visible;">
          {JSON.stringify(resp.body, null, 2)}</pre>
      {:else if resp || !get}
        {@render children?.(resp)}
      {:else}
        <T id="Nodal.noResponse" />
      {/if}
    </ModalBody>

    {#if footer}<ModalFooter>{@render footer(resp)}</ModalFooter>{/if}
  </form>
</Modal>

<style>
  .pre-wrap {
    white-space: pre-wrap;
    word-break: break-all;
    word-wrap: break-word;
  }
</style>
