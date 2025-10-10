<script lang="ts">
  import { Button } from '@sveltestrap/sveltestrap'
  import Fa from '../Fa.svelte'
  import { faFolderOpen } from '@fortawesome/sharp-duotone-regular-svg-icons'
  type Props = { onclick?: () => void; isOpen?: boolean }
  let { onclick, isOpen = $bindable(false) }: Props = $props()
  const click = () => (onclick?.(), (isOpen = true))
</script>

<Button outline onclick={click}>
  <Fa
    i={faFolderOpen}
    scale={1.5}
    c1="Gold"
    c2="DarkGoldenRod"
    d1="Moccasin"
    d2="LemonChiffon" />
</Button>
