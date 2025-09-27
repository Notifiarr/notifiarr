<!-- Displays a simple Languages help modal. -->
<script lang="ts" module>
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
  import ModalWrap from './ModalWrap.svelte'
  import T from '../../includes/Translate.svelte'

  let modal = $state<ModalWrap>()
  export const toggle = () => modal?.toggle()
</script>

<ModalWrap {page} bind:this={modal}><T id="{page.id}.body" /></ModalWrap>
