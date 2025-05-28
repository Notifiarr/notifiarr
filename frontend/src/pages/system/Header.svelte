<script lang="ts">
  import {
    faQuestion,
    type IconDefinition,
  } from '@fortawesome/sharp-duotone-light-svg-icons'
  import Fa, { type Props as FaProps } from '../../includes/Fa.svelte'
  import { _ } from '../../includes/Translate.svelte'

  type Props = { id: string; logo?: string; i?: IconDefinition } & Omit<FaProps, 'i'>
  let {
    id,
    logo = undefined,
    i = faQuestion,
    scale = 1.3,
    style = 'margin-bottom: 3px;',
    ...rest
  }: Props = $props()
</script>

<h4>
  {#if logo}
    <img src={logo} alt="Logo" class="logo" />
  {:else}
    <Fa {...rest} {i} {scale} {style} class="me-2" />
  {/if}
  {$_('system.' + id + '.title')}
</h4>

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
