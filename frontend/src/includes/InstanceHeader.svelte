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

{#snippet logo()}
  {#if typeof flt.app.logo === 'string'}
    <img src={flt.app.logo} alt="Logo" class="logo" />
  {:else}
    <Fa i={flt.app.logo} {...flt.app.iconProps} class="logo" />
  {/if}
{/snippet}

<h4 class="instance-header">
  {#if page}<go-to {page}>{@render logo()}</go-to>{:else}{@render logo()}{/if}

  <T id={flt.app.id + '.title'} />
  {#if flt.removed.length > 0}
    <Badge color="warning" class="ms-3">
      <T id="phrases.DeletedNumber" number={flt.removed.length} />
    </Badge>
  {:else if flt.formChanged}
    <Badge color="warning" class="ms-3"><T id="phrases.Changed" /></Badge>
  {/if}
</h4>

{#if $_(flt.app.id + '.description') !== flt.app.id + '.description'}
  <p><T id={flt.app.id + '.description'} /></p>
{/if}

<style>
  .instance-header :global(.logo) {
    height: 36px;
    margin-right: 6px;
    margin-left: -5px;
    padding-left: 0px;
    vertical-align: bottom;
    display: inline-block;
  }

  .instance-header :global(.badge) {
    font-size: 11px;
    vertical-align: bottom;
    text-transform: none;
    vertical-align: top;
    border-radius: 12px;
    text-align: center;
  }
</style>
