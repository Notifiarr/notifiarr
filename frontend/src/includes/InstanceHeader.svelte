<script lang="ts">
  import { Badge } from '@sveltestrap/sveltestrap'
  import T, { _ } from './Translate.svelte'
  import type { FormListTracker } from './formsTracker.svelte'
  import Fa from './Fa.svelte'

  type Props = {
    flt: FormListTracker<any>
    /** If a page is provided, the header icon gets wrapped in link. */
    page?: string
  }
  let { flt, page = '' }: Props = $props()
</script>

<h4 class="instance-header">
  <Fa i={flt.app.logo} {...flt.app.iconProps} logo page={page || undefined}>
    <T id={flt.app.id + '.title'} />
    {#if flt.removed.length > 0}
      <Badge color="warning" class="ms-3">
        <T id="phrases.DeletedNumber" number={flt.removed.length} />
      </Badge>
    {:else if flt.formChanged}
      <Badge color="warning" class="ms-3"><T id="phrases.Changed" /></Badge>
    {/if}
  </Fa>
</h4>

{#if $_(flt.app.id + '.description') !== flt.app.id + '.description'}
  <p><T id={flt.app.id + '.description'} /></p>
{/if}

<style>
  .instance-header :global(.badge) {
    font-size: 11px;
    text-transform: none;
    border-radius: 12px;
    text-align: center;
  }
</style>
