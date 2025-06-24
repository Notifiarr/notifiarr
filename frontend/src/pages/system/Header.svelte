<script lang="ts">
  import {
    faQuestion,
    type IconDefinition,
  } from '@fortawesome/sharp-duotone-light-svg-icons'
  import Fa, { type Props as FaP } from '../../includes/Fa.svelte'
  import { _ } from '../../includes/Translate.svelte'

  type faProps = Omit<FaP, 'i'>
  type Props = { id: string; logo?: string; i?: IconDefinition; page?: string } & faProps

  let {
    id,
    logo = undefined,
    i = faQuestion,
    scale = 1.3,
    style = 'margin-bottom: 3px;',
    page = undefined,
    ...rest
  }: Props = $props()
</script>

<h4>
  {#if logo}
    <img src={logo} alt="Logo" class="logo" />
  {:else if page}
    <go-to {page}><Fa {...rest} {i} {scale} {style} class="me-2" /></go-to>
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
