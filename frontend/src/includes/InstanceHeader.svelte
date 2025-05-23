<script lang="ts">
  import { Badge } from '@sveltestrap/sveltestrap'
  import T, { _ } from './Translate.svelte'
  import type { App } from './Instance.svelte'

  type Props = { app: App; deleted?: number; changed: boolean }
  let { app, deleted = 0, changed }: Props = $props()
</script>

<h4 class="instance-header">
  <img src={app.logo} alt="Logo" class="logo" />
  <T id={app.id + '.title'} />
  {#if deleted > 0}
    <Badge color="warning" class="ms-3">
      <T id="phrases.DeletedNumber" number={deleted} />
    </Badge>
  {:else if changed}
    <Badge color="warning" class="ms-3"><T id="phrases.Changed" /></Badge>
  {/if}
</h4>

{#if $_(app.id + '.description') !== app.id + '.description'}
  <p><T id={app.id + '.description'} /></p>
{/if}

<style>
  .logo {
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
