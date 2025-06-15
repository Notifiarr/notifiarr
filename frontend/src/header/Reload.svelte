<script lang="ts" module>
  import { _ } from '../includes/Translate.svelte'
  import { get } from 'svelte/store'
  import { getUi, checkReloaded } from '../api/fetch'
  import { updateBackend, showMsg } from './Index.svelte'
  import {
    Button,
    Modal,
    ModalBody,
    ModalFooter,
    ModalHeader,
  } from '@sveltestrap/sveltestrap'
  import { Spinner } from '@sveltestrap/sveltestrap'
  import { theme } from '../includes/theme.svelte'
  import { faRotate } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from '../includes/Fa.svelte'

  let isOpen = $state(false)
  let reloading = $state(false)
  let finish: (() => void) | null = $state(null)

  // Called by the trigger page. Can be used externally to pop up the reload modal.
  export async function reload(e?: Event) {
    e?.preventDefault()
    isOpen = true
    // wait for the modal to be closed.
    await new Promise<void>(resolve => (finish = resolve))
  }

  async function onclick(e?: Event) {
    e?.preventDefault()
    reloading = true

    try {
      await getUi('reload', false) // reload
      await checkReloaded() // wait for reload
    } catch (err) {
      showMsg(
        `<span class="text-danger">
        ${get(_)('phrases.FailedToReload', { values: { error: `${err}` } })}
        </span>`,
      )
    } finally {
      reset(e)
    }

    await updateBackend()
  }

  // Called on cancel and after a reload.
  const reset = (e?: Event) => {
    e?.preventDefault()
    finish?.() // resolve (external) promise
    finish = null // reset (external) promise
    isOpen = false // close the modal
    reloading = false // reset modal state
  }
</script>

<a href="#reload" onclick={e => (e.preventDefault(), (isOpen = true))}>
  <Fa i={faRotate} c1="#33A000" c2="#33A5A4" class="me-2" />
</a>

<Modal {isOpen} toggle={reset} theme={$theme}>
  <ModalHeader>{$_('phrases.ConfirmReload')}</ModalHeader>
  {#if reloading}
    <ModalBody><Spinner size="sm" /> {$_('phrases.Reloading')}</ModalBody>
  {:else}
    <ModalBody>{$_('phrases.ConfirmReloadBody')}</ModalBody>
    <ModalFooter>
      <Button color="danger" {onclick}>{$_('buttons.Confirm')}</Button>
      <Button color="secondary" onclick={reset}>{$_('buttons.Cancel')}</Button>
    </ModalFooter>
  {/if}
</Modal>
