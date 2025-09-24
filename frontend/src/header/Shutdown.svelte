<script lang="ts">
  import { Button, ModalBody, ModalFooter, ModalHeader } from '@sveltestrap/sveltestrap'
  import { getUi } from '../api/fetch'
  import { _ } from '../includes/Translate.svelte'
  import { faPowerOff } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from '../includes/Fa.svelte'
  import MyModal from '../includes/MyModal.svelte'

  let isOpen = $state(false)
  let shutdown: any = $state()
  const onclick = () => (shutdown = async () => await getUi('shutdown', false))
</script>

<a href="#shutdown" onclick={e => (e.preventDefault(), (isOpen = true))} title="shutdown">
  <Fa i={faPowerOff} c1="salmon" c2="maroon" d1="firebrick" d2="red" class="me-2" />
</a>

<MyModal {isOpen} toggle={() => (isOpen = false)}>
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
      <Button color="danger" {onclick}>{$_('buttons.Confirm')}</Button>
      <Button color="secondary" onclick={() => (isOpen = false)}>
        {$_('buttons.Cancel')}</Button>
    </ModalFooter>
  {/if}
</MyModal>
