<script lang="ts">
  import {
    Button,
    Modal,
    ModalBody,
    ModalFooter,
    ModalHeader,
  } from '@sveltestrap/sveltestrap'
  import { getUi, checkReloaded } from '../api/fetch'
  import { _ } from '../includes/Translate.svelte'
  import { Spinner } from '@sveltestrap/sveltestrap'
  import { theme } from '../includes/theme.svelte'
  import { updateBackend, showMsg } from './Index.svelte'
  import { faRotate } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from '../includes/Fa.svelte'

  let isOpen = $state(false)
  let reloading = $state(false)

  async function onclick(e?: Event) {
    e?.preventDefault()
    reloading = true

    try {
      await getUi('reload', false) // reload
      await checkReloaded() // wait for reload
    } catch (err) {
      showMsg(
        `<span class="text-danger">
        ${$_('phrases.FailedToReload', { values: { error: `${err}` } })}
        </span>`,
      )
    } finally {
      isOpen = false // close the modal
      reloading = false // reset modal state
    }

    await updateBackend()
  }
</script>

<a href="#reload" onclick={e => (e.preventDefault(), (isOpen = true))}>
  <Fa i={faRotate} c1="#33A000" c2="#33A5A4" class="me-2" />
</a>

<Modal {isOpen} toggle={() => (isOpen = false)} theme={$theme}>
  <ModalHeader>{$_('phrases.ConfirmReload')}</ModalHeader>
  {#if reloading}
    <ModalBody><Spinner size="sm" /> {$_('phrases.Reloading')}</ModalBody>
  {:else}
    <ModalBody>{$_('phrases.ConfirmReloadBody')}</ModalBody>
    <ModalFooter>
      <Button color="danger" {onclick}>{$_('buttons.Confirm')}</Button>
      <Button color="secondary" onclick={() => (isOpen = false)}>
        {$_('buttons.Cancel')}</Button>
    </ModalFooter>
  {/if}
</Modal>
