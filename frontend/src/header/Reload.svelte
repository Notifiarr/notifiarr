<script lang="ts" module>
  import { _ } from '../includes/Translate.svelte'
  import { get } from 'svelte/store'
  import { getUi, checkReloaded } from '../api/fetch'
  import { updateBackend, showMsg } from './Index.svelte'
  import { Button } from '@sveltestrap/sveltestrap'
  import { Spinner } from '@sveltestrap/sveltestrap'
  import { faRotate } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from '../includes/Fa.svelte'
  import Nodal from '../includes/Nodal.svelte'

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
      reset()
    }

    await updateBackend()
  }

  // Called on cancel and after a reload.
  const reset = () => {
    finish?.() // resolve (external) promise
    finish = null // reset (external) promise
    isOpen = false // close the modal
    reloading = false // reset modal state
  }
</script>

<a href="#reload" onclick={e => (e.preventDefault(), (isOpen = true))} title="reload">
  <Fa i={faRotate} c1="#33A000" c2="#33A5A4" class="me-2" spin={isOpen} />
</a>

<Nodal bind:isOpen title="phrases.ConfirmReload" follow={reset} disabled={reloading} esc>
  {#if reloading}
    <Spinner size="sm" /> {$_('phrases.Reloading')}
  {:else}
    {$_('phrases.ConfirmReloadBody')}
  {/if}
  {#snippet footer()}
    <Button color="danger" {onclick} disabled={reloading}>
      {$_('buttons.Confirm')}</Button>
    <Button color="secondary" onclick={reset} disabled={reloading}>
      {$_('buttons.Cancel')}</Button>
  {/snippet}
</Nodal>
