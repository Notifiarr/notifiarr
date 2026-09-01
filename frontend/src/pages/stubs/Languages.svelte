<!-- Displays a simple Languages help modal. -->
<script lang="ts" module>
  import Nodal from '../../includes/Nodal.svelte'
  import T from '../../includes/Translate.svelte'

  export const page = {
    type: 'modal' as const,
    id: 'Languages',
  }
</script>

<script lang="ts">
  let isOpen = $state(false)
  export const toggle = () => (isOpen = !isOpen)
</script>

<Nodal bind:isOpen title={page.id} size="lg" esc>
  <T id="{page.id}.body" /></Nodal>
