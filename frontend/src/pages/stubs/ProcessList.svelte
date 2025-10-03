<!-- Process list page: ps aux -->
<script lang="ts" module>
  import { getUi } from '../../api/fetch'
  import { Button } from '@sveltestrap/sveltestrap'
  import T from '../../includes/Translate.svelte'
  import Nodal from '../../includes/Nodal.svelte'
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
  let wrap = $state(false)
  let isOpen = $state(false)
  const get = async () => await getUi('ps', false)
  export const toggle = () => (isOpen = !isOpen)
</script>

<Nodal {get} bind:isOpen title={page.id} fa={page} esc size="xl" full>
  {#snippet children(resp)}
    <pre style="overflow: visible;" class:wrap>{resp?.body ?? ''}</pre>
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
</Nodal>

<style>
  pre.wrap {
    white-space: pre-wrap;
    word-break: break-all;
    word-wrap: break-word;
  }
</style>
