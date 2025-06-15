<script lang="ts">
  import {
    CardHeader,
    Dropdown,
    DropdownItem,
    DropdownMenu,
    DropdownToggle,
    Popover,
    Table,
  } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import Fa from '../../includes/Fa.svelte'
  import { faCirclePlay } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { type TriggerInfo } from '../../api/notifiarrConfig'
  import { theme as thm } from '../../includes/theme.svelte'
  import { profile } from '../../api/profile.svelte'
  import { run } from './run'
  import { failure, successIf } from '../../includes/util'

  type Props = { type: 'Triggers' | 'Timers' | 'Schedules'; info: TriggerInfo }
  let { type, info }: Props = $props()

  const id = btoa(info.key + info.name).split('=')[0]
  const theme = $derived($thm)
  let disabled = $state(false)

  const onclick = async (e: Event, option?: any) => {
    e.preventDefault()
    if (knownOptions[info.key] && !option) return
    disabled = true
    const resp = await run(info, option)
    if (resp.ok) successIf('Actions.triggers.' + info.key + '.success')
    else failure(resp.body)
    disabled = false
  }

  const b = 'Actions.triggers.' + info.key + '.button'
  const button = $derived($_(b) != b ? $_(b) : '')

  const knownOptions: Record<string, string[]> = {
    TrigUploadFile: ['app', 'debug', 'http'],
  }
</script>

<CardHeader>
  <Popover target={id} trigger="hover" {theme}>
    {#if info.key == 'TrigCustomCronTimer'}
      {$profile.siteCrons?.find(c => info.name.endsWith("'" + c.name + "'"))?.description}
    {:else}
      {info.name}
    {/if}
  </Popover>
  <div {id}><T id="Actions.triggers.{info.key}.label" name={info.name} /></div>
</CardHeader>

<Table class="mb-0" borderless striped>
  <tbody class="fit">
    <tr>
      <th><T id="Actions.titles.Counter" /></th>
      <td>
        {info.runs || 0}
        <!-- If a button is defined in the translation file, show a button to trigger the action. -->
        {#if button}
          <Popover target={id + 'b'} trigger="hover" {theme}>{button}</Popover>
          <Dropdown class="d-inline-block float-end">
            <DropdownToggle id={id + 'b'} {onclick} {disabled} color="transparent">
              <Fa
                i={faCirclePlay}
                spin={disabled}
                scale={1.3}
                c1="green"
                c2="wheat"
                d1="limegreen"
                d2="darkcyan" />
            </DropdownToggle>
            <!-- If there are options (defined above), show a dropdown to select one. -->
            {#if knownOptions[info.key]}
              <DropdownMenu>
                <DropdownItem header>
                  {$_('Actions.triggers.' + info.key + '.options.choose')}
                </DropdownItem>
                {#each knownOptions[info.key] as option}
                  <DropdownItem onclick={e => onclick(e, option)}>
                    {$_('Actions.triggers.' + info.key + '.options.' + option)}
                  </DropdownItem>
                {/each}
              </DropdownMenu>
            {/if}
          </Dropdown>
        {/if}
      </td>
    </tr>

    {#if type === 'Timers'}
      <tr><th><T id="Actions.titles.Interval" /></th><td>{info.dur}</td></tr>
    {:else if type === 'Schedules'}
      <tr><th><T id="Actions.titles.Schedule" /></th><td>{info.dur}</td></tr>
    {/if}
  </tbody>
</Table>

<style>
  tbody.fit th {
    width: 1%;
    white-space: nowrap;
    margin-right: 1rem;
  }

  tbody.fit td {
    text-align: left;
  }

  tbody.fit td :global(button) {
    padding: 0 0.25rem !important;
    margin-left: 0.5rem !important;
    margin-bottom: 0 !important;
    max-height: 25px !important;
  }
</style>
