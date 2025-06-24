<!-- Process list page: ps aux -->
<script lang="ts" module>
  import { faListTree } from '@fortawesome/sharp-duotone-solid-svg-icons'
  export const page = {
    type: 'modal' as const,
    id: 'ProcessList',
    i: faListTree,
    c1: 'coral',
    c2: 'steelblue',
    d1: 'wheat',
    d2: 'lime',
  }
</script>

<script lang="ts">
  import { getUi } from '../../api/fetch'
  import { Button } from '@sveltestrap/sveltestrap'
  import T from '../../includes/Translate.svelte'
  import ModalWrap from './ModalWrap.svelte'

  let wrap = $state(false)
  let isOpen = $state(false)
  const get = async () => await getUi('ps', false)
  export const toggle = () => (isOpen = !isOpen)
</script>

<ModalWrap {page} {get} bind:isOpen>
  {#snippet children(ps)}
    <pre style="overflow: visible;" class:wrap>{ps}</pre>
  {/snippet}
  {#snippet footer(resp)}
    {#if resp && resp.ok}
      <small class="text-muted">
        <T id="{page.id}.footer" count={resp.body.split('\n').length} />
      </small>
    {/if}
    <Button outline color="warning" size="sm" on:click={() => (wrap = !wrap)}>
      <T id="{page.id}.button.{wrap ? 'wrapOff' : 'wrapOn'}" />
    </Button>
  {/snippet}
</ModalWrap>

<style>
  pre.wrap {
    white-space: pre-wrap;
    word-break: break-all;
    word-wrap: break-word;
  }
</style>
