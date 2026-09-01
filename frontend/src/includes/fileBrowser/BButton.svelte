<script lang="ts">
  import { Button } from '@sveltestrap/sveltestrap'
  import Fa from '../Fa.svelte'
  import FolderOpen from 'phosphor-svelte/lib/FolderOpen'
  import { _ } from '../Translate.svelte'
  type Props = { onclick?: () => void; isOpen?: boolean }
  let { onclick, isOpen = $bindable(false) }: Props = $props()
  const click = () => (onclick?.(), (isOpen = true))
</script>

<Button
  type="button"
  outline
  onclick={click}
  style="width:44px;"
  title={$_('phrases.BrowseFiles')}
  aria-label={$_('phrases.BrowseFiles')}>
  <Fa i={FolderOpen} btn c1="Gold" d1="Moccasin" />
</Button>
