<!-- Displays the client info payload from the website as JSON in a modal. -->
<script lang="ts" module>
  import Nodal from '../../includes/Nodal.svelte'
  import { profile } from '../../api/profile.svelte'
  import T from '../../includes/Translate.svelte'

  export const page = {
    type: 'modal' as const,
    id: 'ClientInfo',
  }
</script>

<script lang="ts">
  let isOpen = $state(false)
  export const toggle = () => (isOpen = !isOpen)
</script>

<Nodal bind:isOpen title={page.id} esc size="xl" full>
  <pre style="overflow: visible;">
      {JSON.stringify($profile.clientInfo, null, 2)}</pre>
  {#snippet footer()}
    <small class="text-muted"><T id="{page.id}.description" /></small>
  {/snippet}
</Nodal>
