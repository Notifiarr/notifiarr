<!-- H4 (or whatever element) abstraction with a logo or icon and a title. -->

<script lang="ts">
  import {
    faQuestion,
    type IconDefinition,
  } from '@fortawesome/sharp-duotone-light-svg-icons'
  import Fa, { type Props as FaProps } from './Fa.svelte'
  import { _, json } from './Translate.svelte'

  type Props = {
    id: string
    logo?: string
    i?: IconDefinition
    page?: string
    parent?: string
    elem?: string
    elemstyle?: string
    class?: string
  } & Omit<FaProps, 'i'>

  let {
    id,
    logo = undefined,
    parent = 'system',
    i = faQuestion,
    scale = 1.3,
    style = 'margin-bottom: 3px;',
    page = undefined,
    elem = 'h4',
    elemstyle = '',
    class: className = '',
    ...rest
  }: Props = $props()

  const title = $json([parent, id].filter(v => v).join('.'))
</script>

{#snippet image()}
  {#if logo}
    <img src={logo} alt="Logo" class="logo" />
  {:else}
    <Fa {...rest} {i} {scale} {style} class="me-2" />
  {/if}
{/snippet}

<svelte:element this={elem} class={className} style={elemstyle}>
  {#if page}
    <go-to {page}>{@render image()}</go-to>
  {:else}
    {@render image()}
  {/if}
  {typeof title === 'string' ? title : (title as any)['title']}
</svelte:element>

<style>
  .logo {
    height: 36px;
    margin-right: 6px;
    margin-left: -5px;
    padding-left: 0px;
    vertical-align: bottom;
    display: inline-block;
  }
</style>
