<script lang="ts">
  import Nodal from '../../includes/Nodal.svelte'
  import {
    faCodeSimple,
    type IconDefinition,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'

  type Props = { data: any; pageId: string; icon?: IconDefinition }
  let { data, pageId, icon = faCodeSimple }: Props = $props()
  let isOpen = $state(false)
  export const toggle = (e?: Event) => (e?.preventDefault?.(), (isOpen = !isOpen))
</script>

<Nodal title={'Integrations.' + pageId} fa={{ i: icon }} bind:isOpen>
  <pre style="overflow: visible;">{JSON.stringify(data, null, 2)}</pre>
</Nodal>
