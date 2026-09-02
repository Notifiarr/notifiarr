<!-- H4 (or whatever element) abstraction with a logo or icon and a title. -->

<script lang="ts">
  import Fa, { type Props as FaProps, type IconSource } from './Fa.svelte'
  import { _, json } from './Translate.svelte'

  type Props = {
    id: string
    /** Image URL shorthand (same as `i` with a PNG/SVG import). Not `Fa.logo`. */
    logo?: string
    i?: IconSource
    page?: string
    parent?: string
    elem?: string
    elemstyle?: string
    class?: string
  } & Omit<FaProps, 'i' | 'logo' | 'page' | 'children'>

  let {
    id,
    logo: src,
    parent = 'system',
    i,
    page,
    elem = 'h4',
    elemstyle = '',
    class: className = '',
    ...rest
  }: Props = $props()

  const title = $derived($json([parent, id].filter(v => v).join('.')))
</script>

<svelte:element this={elem} class={['helem', className]} style={elemstyle}>
  <Fa {...rest} i={src ?? i} {page} logo>
    {typeof title === 'string' ? title : (title as any)['title']}
  </Fa>
</svelte:element>
