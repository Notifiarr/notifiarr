<script lang="ts">
  import { type BackendResponse } from '../../api/fetch'
  import {
    Button,
    ButtonGroup,
    Modal,
    ModalBody,
    ModalFooter,
    ModalHeader,
    Popover,
  } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import Fa from '../../includes/Fa.svelte'
  import type { Page } from '../../navigation/nav.svelte'
  import type { Snippet } from 'svelte'
  import {
    faArrowsRotate,
    faCompress,
    faExpand,
    faXmarkLarge,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { delay } from '../../includes/util'

  type Props = {
    children: Snippet<[any]>
    page: Omit<Page, 'component'>
    get?: () => Promise<BackendResponse>
    footer?: Snippet<[any]>
    isOpen: boolean
  }

  let { children, page, get, footer, isOpen = $bindable(false) }: Props = $props()
  let loading = $state(false)
  let resp = $state<BackendResponse>()
  let fullscreen = $state(false)
  export const toggle = () => (isOpen = !isOpen)
  const height = $derived(footer ? 'calc(100vh - 180px)' : 'calc(100vh - 110px)')

  const refresh = async () => {
    try {
      loading = true
      resp = get ? await get() : undefined
      await delay(300)
    } catch (error) {
      resp = { ok: false, body: error }
    } finally {
      loading = false
    }
  }

  $effect(() => {
    if (isOpen) refresh()
  })
</script>

<Modal {isOpen} size="xl" theme={$theme} {fullscreen} {toggle}>
  <ModalHeader class="d-inline-block">
    <Fa {...page} scale={1.4} class="me-2" />
    <T id="{page.id}.title" />

    <ButtonGroup class="float-end">
      {#if get}
        <Button
          id="refreshM"
          outline
          color="secondary"
          size="sm"
          on:click={refresh}
          aria-label={$_('ModalWrap.button.refresh')}
          title={$_('ModalWrap.button.refresh')}>
          <Fa
            i={faArrowsRotate}
            c1="steelblue"
            c2="firebrick"
            d2="pink"
            scale={1.5}
            spin={loading} />
        </Button>
      {/if}
      <Button
        id="fullscreenM"
        outline
        color="secondary"
        size="sm"
        on:click={() => (fullscreen = !fullscreen)}
        aria-label={$_('ModalWrap.button.fullscreen')}
        title={$_('ModalWrap.button.fullscreen')}>
        <Fa
          i={fullscreen ? faCompress : faExpand}
          c1="steelblue"
          c2="firebrick"
          d2="pink"
          scale={1.5} />
      </Button>
      <Button
        id="closeM"
        outline
        color="secondary"
        size="sm"
        title={$_('buttons.Close')}
        aria-label={$_('buttons.Close')}
        on:click={() => (isOpen = false)}>
        <Fa i={faXmarkLarge} c2="orange" d2="gold" scale={1.5} />
      </Button>
    </ButtonGroup>

    <Popover target="refreshM" trigger="hover" theme={$theme}>
      <T id="ModalWrap.button.refresh" />
    </Popover>
    <Popover target="fullscreenM" trigger="hover" theme={$theme}>
      <T id="ModalWrap.button.fullscreen" />
    </Popover>
    <Popover target="closeM" trigger="hover" theme={$theme}>
      <T id="buttons.Close" />
    </Popover>
  </ModalHeader>

  <ModalBody style="max-height: {height}; overflow: auto;">
    {#if loading && !resp}
      <T id="phrases.Loading" />
    {:else if (resp && resp.ok) || !get}
      {@render children(resp?.body)}
    {:else if resp && !resp.ok}
      <T id="{page.id}.error" error={resp.body} />
    {:else}
      <T id="{page.id}.noData" />
    {/if}
  </ModalBody>

  {#if footer}<ModalFooter>{@render footer(resp)}</ModalFooter>{/if}
</Modal>
