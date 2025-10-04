<script lang="ts">
  import { Popover } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import { theme as thm } from '../../includes/theme.svelte'
  import T from '../../includes/Translate.svelte'
  import type { TriggerInfo } from '../../api/notifiarrConfig'

  type Props = { row: TriggerInfo }
  const { row }: Props = $props()
  const id = $derived(btoa(row.key.split('').reverse().join('') + row.name).split('=')[0])
</script>

<!-- We use the row.name in a popover. It's not translated. -->
<Popover target={id + 'label'} trigger="hover" theme={$thm}>
  <code>{row.key}</code><br />
  {#if row.key == 'TrigCustomCronTimer'}
    {$profile.siteCrons?.find(c => row.name.endsWith("'" + c.name + "'"))?.description}
  {:else}
    {row.name}
  {/if}
</Popover>

<!-- We use the row.key to translate the name. -->
<span id={id + 'label'}>
  {#if row.key == 'TrigCustomCronTimer'}
    <T id="Actions.triggers.{row.key}.label" name={row.name.split("'")[1]} />
  {:else}
    <T id="Actions.triggers.{row.key}.label" name={row.name} />
  {/if}
</span>
