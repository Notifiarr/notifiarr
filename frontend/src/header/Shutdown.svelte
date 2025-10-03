<script lang="ts">
  import { Button } from '@sveltestrap/sveltestrap'
  import { getUi } from '../api/fetch'
  import { _ } from '../includes/Translate.svelte'
  import { faPowerOff } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from '../includes/Fa.svelte'
  import Nodal from '../includes/Nodal.svelte'

  let isOpen = $state(false)
  let shutdown: any = $state()
  const onclick = () => (shutdown = async () => await getUi('shutdown', false))
</script>

<a href="#shutdown" onclick={e => (e.preventDefault(), (isOpen = true))} title="shutdown">
  <Fa i={faPowerOff} c1="salmon" c2="maroon" d1="firebrick" d2="red" class="me-2" />
</a>

<Nodal bind:isOpen title="phrases.ConfirmShutdown" disabled={shutdown} esc>
  {#if shutdown}
    {#await shutdown() then result}
      {#if result.ok}
        <span class="text-danger">{$_('phrases.ShutdownSuccess')}</span>
      {:else}
        {$_('phrases.FailedToShutdown', { values: { error: result.body } })}
      {/if}
    {/await}
  {:else}
    {$_('phrases.ConfirmShutdownBody')}
  {/if}
  {#snippet footer()}
    <Button color="danger" {onclick} disabled={shutdown}>
      {$_('buttons.Confirm')}</Button>
    <Button color="secondary" onclick={() => (isOpen = false)} disabled={shutdown}>
      {$_('buttons.Cancel')}</Button>
  {/snippet}
</Nodal>
