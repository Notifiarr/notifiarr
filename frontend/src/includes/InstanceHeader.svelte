<script lang="ts">
  import { Badge } from '@sveltestrap/sveltestrap'
  import T, { _ } from './Translate.svelte'
  import type { InstanceFormValidator } from './instanceFormValidator.svelte'

  type Props = { iv: InstanceFormValidator }
  let { iv }: Props = $props()
</script>

<h4 class="instance-header">
  <img src={iv.app.logo} alt="Logo" class="logo" />
  <T id={iv.app.id + '.title'} />
  {#if iv.removed.length > 0}
    <Badge color="warning" class="ms-3">
      <T id="phrases.DeletedNumber" number={iv.removed.length} />
    </Badge>
  {:else if iv.formChanged}
    <Badge color="warning" class="ms-3"><T id="phrases.Changed" /></Badge>
  {/if}
</h4>

{#if $_(iv.app.id + '.description') !== iv.app.id + '.description'}
  <p><T id={iv.app.id + '.description'} /></p>
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
