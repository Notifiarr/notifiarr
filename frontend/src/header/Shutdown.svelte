<script lang="ts">
  import {
    Button,
    Modal,
    ModalBody,
    ModalFooter,
    ModalHeader,
  } from '@sveltestrap/sveltestrap'
  import { getUi } from '../api/fetch'
  import { theme } from '../includes/theme.svelte'
  import { _ } from '../includes/Translate.svelte'

  type Props = { isOpen: boolean; toggle: () => void }
  const { isOpen = $bindable(false), toggle = $bindable(() => {}) }: Props = $props()
  let shutdown: any = $state()
</script>

<Modal {isOpen} {toggle} theme={$theme}>
  <ModalHeader>{$_('phrases.ConfirmShutdown')}</ModalHeader>
  {#if shutdown}
    <ModalBody>
      {#await shutdown() then result}
        {#if result.ok}
          <span class="text-danger">{$_('phrases.ShutdownSuccess')}</span>
        {:else}
          {$_('phrases.FailedToShutdown', { values: { error: result.body } })}
        {/if}
      {/await}
    </ModalBody>
  {:else}
    <ModalBody>{$_('phrases.ConfirmShutdownBody')}</ModalBody>
    <ModalFooter>
      <Button
        color="danger"
        onclick={() => (shutdown = async () => await getUi('shutdown', false))}>
        {$_('buttons.Confirm')}</Button>
      <Button color="secondary" onclick={toggle}>
        {$_('buttons.Cancel')}</Button>
    </ModalFooter>
  {/if}
</Modal>
