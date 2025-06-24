<!-- Process list page: ps aux -->
<script lang="ts" module>
  import { faCodeCompare } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    type: 'modal' as const,
    id: 'ClientInfo',
    i: faCodeCompare,
    c1: 'coral',
    c2: 'steelblue',
    d1: 'lime',
    d2: 'wheat',
  }
</script>

<script lang="ts">
  import { type BackendResponse } from '../../api/fetch'
  import ModalWrap from './ModalWrap.svelte'
  import { profile } from '../../api/profile.svelte'
  import T from '../../includes/Translate.svelte'

  let isOpen = $state(false)
  const get = async () =>
    ({
      ok: true,
      body: await JSON.stringify($profile.clientInfo, null, 2),
    }) as BackendResponse
  export const toggle = () => (isOpen = !isOpen)
</script>

<ModalWrap {page} {get} bind:isOpen>
  {#snippet children(ps)}
    <pre style="overflow: visible;">{ps}</pre>
  {/snippet}
  {#snippet footer()}
    <small class="text-muted"><T id="{page.id}.description" /></small>
  {/snippet}
</ModalWrap>
