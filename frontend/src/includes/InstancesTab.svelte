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
  import type { InstanceFormValidator } from './instanceFormValidator.svelte'

  type Props = { iv: InstanceFormValidator; tab: string; titles: Record<string, string> }
  let { iv, tab = $bindable(), titles = $bindable() }: Props = $props()
</script>

<TabPane
  tabId={iv.app.name.toLowerCase()}
  active={tab === iv.app.name.toLowerCase()}
  onclick={() => (tab = iv.app.name.toLowerCase())}>
  <div class="title" slot="tab">
    <h5 class="title {iv.invalid ? 'text-danger' : ''}">
      {titles[iv.app.name]}
      {#if iv.instances.length > 0}
        <Badge class="tab-badge" color="success">{iv.instances.length}</Badge>
      {/if}
      {#if iv.invalid}
        <Fa i={faExclamationTriangle} c1="red" c2="red" />
      {:else if iv.formChanged || iv.removed.length > 0}
        <Fa i={faExclamationTriangle} c1="orange" c2="orange" />
      {/if}
    </h5>
  </div>

  <Instances {iv} />
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
