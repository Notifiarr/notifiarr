<script lang="ts">
  import Nodal from '../Nodal.svelte'
  import Browser from './Index.svelte'

  let {
    title = 'File Browser',
    isOpen = $bindable(false),
    value = $bindable(''),
    description = '',
    ...rest
  }: {
    isOpen: boolean
    value: string
    title: string
    description: string
    [key: string]: any
  } = $props()

  const close = () => (isOpen = false)
  let fullscreen = $state(false)
  let maxHeight = $derived('max-height: ' + (fullscreen ? '100vh' : '1000px'))
  let height = $derived('calc(100vh - ' + (fullscreen ? '100' : '155') + 'px);')
</script>

<Nodal bind:isOpen size="lg" bind:fullscreen {title}>
  <Browser bind:value height="{height};{maxHeight}" {close} {description} {...rest} />
</Nodal>
