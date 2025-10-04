<script lang="ts">
  import {
    Dropdown,
    DropdownItem,
    DropdownMenu,
    DropdownToggle,
    Popover,
  } from '@sveltestrap/sveltestrap'
  import { theme as thm } from '../../includes/theme.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import Fa from '../../includes/Fa.svelte'
  import { run, type Row } from './run'
  import { failure, successIf } from '../../includes/util'
  import { faCirclePlay } from '@fortawesome/sharp-duotone-light-svg-icons'

  type Props = { row: Row }
  const { row }: Props = $props()

  const id = $derived(btoa(row.key.split('').reverse().join('') + row.name).split('=')[0])
  const b = $derived('Actions.triggers.' + row.key + '.button')
  const button = $derived($_(b) != b ? $_(b) : '')
  let disabled = $state(false)

  const knownOptions: Record<string, string[]> = {
    TrigUploadFile: ['app', 'debug', 'http'],
  }

  const onclick = async (e: Event, option?: any) => {
    e.preventDefault()
    if (knownOptions[row.key] && !option) return
    disabled = true
    const resp = await run(row, option)
    if (resp.ok) successIf('Actions.triggers.' + row.key + '.success')
    else failure(resp.body)
    disabled = false
  }
</script>

<div class="d-inline-block">{row.runs || 0}</div>
<!-- If a button is defined in the translation file, show a button to trigger the action. -->
<div id="runs-cell" class="float-end">
  {#if button}
    <Dropdown>
      <Popover target={id + 'button'} trigger="hover" theme={$thm}>{button}</Popover>
      <DropdownToggle
        id={id + 'button'}
        {onclick}
        {disabled}
        color="transparent"
        class="p-0 m-0 ms-1"
        size="sm">
        <Fa
          style="vertical-align: text-top;"
          i={faCirclePlay}
          spin={disabled}
          scale={1.2}
          c1="green"
          c2="wheat"
          d1="limegreen"
          d2="darkcyan" />
      </DropdownToggle>
      <!-- If there are options (defined above), show a dropdown to select one. -->
      {#if knownOptions[row.key]}
        <DropdownMenu>
          <DropdownItem header>
            {$_('Actions.triggers.' + row.key + '.options.choose')}
          </DropdownItem>
          {#each knownOptions[row.key] as option}
            <DropdownItem onclick={e => onclick(e, option)}>
              {$_('Actions.triggers.' + row.key + '.options.' + option)}
            </DropdownItem>
          {/each}
        </DropdownMenu>
      {/if}
    </Dropdown>
  {/if}
</div>
