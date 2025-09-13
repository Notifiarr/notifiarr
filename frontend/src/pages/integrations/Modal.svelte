<script lang="ts">
  import ModalWrap from '../stubs/ModalWrap.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import {
    faCodeSimple,
    type IconDefinition,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import type { Snippet } from 'svelte'

  type Props = { children?: Snippet; data?: any; pageId: string; icon?: IconDefinition }
  let { children, data, pageId, icon = faCodeSimple }: Props = $props()

  let isOpen = $state(false)
  export const toggle = (e?: Event) => (e?.preventDefault?.(), (isOpen = !isOpen))
</script>

<ModalWrap page={{ id: 'Integrations.' + pageId, i: icon }} bind:isOpen>
  {#if data}
    <pre style="overflow: visible;">{JSON.stringify(data, null, 2)}</pre>
  {:else if children}
    {@render children()}
  {/if}
</ModalWrap>
