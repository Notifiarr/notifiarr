<!-- Displays a simple Languages help modal. -->
<script lang="ts" module>
  import Nodal from '../../includes/Nodal.svelte'
  import T from '../../includes/Translate.svelte'
  import { faLanguage } from '@fortawesome/sharp-duotone-light-svg-icons'

  export const page = {
    type: 'modal' as const,
    id: 'Languages',
    i: faLanguage,
    c1: 'coral',
    c2: 'steelblue',
    d1: 'darkcyan',
    d2: 'darkslategray',
  }
</script>

<script lang="ts">
  let isOpen = $state(false)
  export const toggle = () => (isOpen = !isOpen)
</script>

<Nodal bind:isOpen title={page.id} fa={page} size="lg" esc>
  <T id="{page.id}.body" /></Nodal>
