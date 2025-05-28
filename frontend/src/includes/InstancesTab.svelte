<script lang="ts" module>
  const uriTab = window.location.pathname.split('/').pop()?.toLowerCase() ?? ''
  export const getTab = (tabs: string[]) =>
    Object.values(tabs).includes(uriTab) ? uriTab : tabs[0]
</script>

<script lang="ts">
  import Fa from './Fa.svelte'
  import { faExclamationTriangle } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { Badge, TabPane } from '@sveltestrap/sveltestrap'
  import Instances from './Instances.svelte'
  import type { FormListTracker } from './formsTracker.svelte'
  import Instance from './Instance.svelte'

  type Props<T> = { flt: FormListTracker<T>; tab: string; titles: Record<string, string> }
  let { flt, tab = $bindable(), titles = $bindable() }: Props<any> = $props()
</script>

<TabPane
  tabId={flt.app.name.toLowerCase()}
  active={tab === flt.app.name.toLowerCase()}
  onclick={() => (tab = flt.app.name.toLowerCase())}>
  <div class="title" slot="tab">
    <h5 class="title {flt.invalid ? 'text-danger' : ''}">
      {titles[flt.app.name]}
      {#if flt.instances.length > 0}
        <Badge class="tab-badge" color="success">{flt.instances.length}</Badge>
      {/if}
      {#if flt.invalid}
        <Fa i={faExclamationTriangle} c1="red" c2="red" />
      {:else if flt.formChanged || flt.removed.length > 0}
        <Fa i={faExclamationTriangle} c1="orange" c2="orange" />
      {/if}
    </h5>
  </div>

  <Instances {flt} Child={Instance}>
    {#snippet headerActive(index)}
      {index + 1}. {flt.original[index]?.name}
    {/snippet}
    {#snippet headerCollapsed(index)}
      {flt.original[index]?.url}
    {/snippet}
  </Instances>
</TabPane>

<style>
  .title {
    display: inline-block;
  }

  .title :global(.tab-badge) {
    padding: 2px 4px 3px;
    font-size: 11px;
    vertical-align: top;
  }
</style>
