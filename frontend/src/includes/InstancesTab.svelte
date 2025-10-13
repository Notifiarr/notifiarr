<script lang="ts" module>
  import { nav } from '../navigation/nav.svelte'
  // This component may only be loaded once per page.
  let tab = $state('')
  /** Set the current tab based on the URI. */
  export const setTab = (tabs: string[]) => {
    const uriTab = window.location.pathname.split('/').pop()?.toLowerCase() ?? ''
    tab = Object.values(tabs).includes(uriTab) ? uriTab : tabs[0]
  }
  /** Goto is a shortcut to nav.goto when the tab changes. */
  export const goto = (e: CustomEvent<string | number>, pid: string) =>
    tab != e.detail && nav.updateURI(pid, [e.detail.toString()])
</script>

<script lang="ts">
  import Fa from './Fa.svelte'
  import { faExclamationTriangle } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { Badge, TabPane } from '@sveltestrap/sveltestrap'
  import Instances from './Instances.svelte'
  import type { FormListTracker } from './formsTracker.svelte'
  import Instance from './Instance.svelte'
  import InstanceHeader from './InstanceHeader.svelte'

  type Props<T> = {
    flt: FormListTracker<T>
    titles: Record<string, string>
    one?: boolean
  }
  let { flt = $bindable(), titles, one = false }: Props<any> = $props()
</script>

{#key tab}
  <TabPane tabId={flt.app.name.toLowerCase()} active={tab === flt.app.name.toLowerCase()}>
    <div class="title" slot="tab">
      <h5 class="title {flt.invalid ? 'text-danger' : ''}">
        {titles[flt.app.name]}
        {#if flt.instances.length > 0 && (!one || !flt.instances[0].disabled)}
          <Badge class="tab-badge" color="success">{flt.instances.length}</Badge>
        {/if}
        {#if flt.invalid}
          <Fa i={faExclamationTriangle} c1="red" c2="red" />
        {:else if flt.formChanged || flt.removed.length > 0}
          <Fa i={faExclamationTriangle} c1="orange" c2="orange" />
        {/if}
      </h5>
    </div>

    {#if one}
      <InstanceHeader {flt} />
      <Instance
        index={0}
        indexed={false}
        bind:form={flt.instances[0]}
        original={flt.original[0]}
        app={flt.app} />
    {:else}
      <Instances bind:flt Child={Instance}>
        {#snippet headerActive(index)}
          {index + 1}. {flt.original[index]?.name}
        {/snippet}
        {#snippet headerCollapsed(index)}
          {flt.original[index]?.url}
        {/snippet}
      </Instances>
    {/if}
  </TabPane>
{/key}

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
