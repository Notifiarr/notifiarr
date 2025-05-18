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

  type Props = { isOpen: boolean; toggle: () => void }
  const { isOpen = $bindable(false), toggle = $bindable(() => {}) }: Props = $props()

  let reloading = $state(false)

  async function reloadBackendConfig(e?: Event) {
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
      toggle() // close the modal
      reloading = false // reset modal state
    }

    await updateBackend()
  }
</script>

<Modal {isOpen} {toggle} theme={$theme}>
  <ModalHeader>{$_('phrases.ConfirmReload')}</ModalHeader>
  {#if reloading}
    <ModalBody><Spinner size="sm" /> {$_('phrases.Reloading')}</ModalBody>
  {:else}
    <ModalBody>{$_('phrases.ConfirmReloadBody')}</ModalBody>
    <ModalFooter>
      <Button color="danger" onclick={reloadBackendConfig}>
        {$_('buttons.Confirm')}</Button>
      <Button color="secondary" onclick={toggle}>
        {$_('buttons.Cancel')}</Button>
    </ModalFooter>
  {/if}
</Modal>
