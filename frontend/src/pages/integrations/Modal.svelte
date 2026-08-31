<script lang="ts">
  import Nodal from '../../includes/Nodal.svelte'
  import CodeSimple from 'phosphor-svelte/lib/CodeSimple'
  import type { IconSource } from '../../includes/Fa.svelte'

  type Props = { data: any; pageId: string; icon?: IconSource }
  let { data, pageId, icon = CodeSimple }: Props = $props()
  let isOpen = $state(false)
  export const toggle = (e?: Event) => (e?.preventDefault?.(), (isOpen = !isOpen))
</script>

<Nodal title={'Integrations.' + pageId} fa={{ i: icon }} bind:isOpen size="xl">
  <pre style="overflow: visible;">{JSON.stringify(data, null, 2)}</pre>
</Nodal>
