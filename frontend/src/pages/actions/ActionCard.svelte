<script lang="ts">
  import { CardHeader, Table } from '@sveltestrap/sveltestrap'
  import T from '../../includes/Translate.svelte'
  import { rtrim } from '../../includes/util'
  import NameCell from './NameCell.svelte'
  import RunsCell from './RunsCell.svelte'
  import { type Row } from './run'

  type Props = { type: 'Triggers' | 'Timers' | 'Schedules'; row: Row }
  const { type, row }: Props = $props()
</script>

<CardHeader><NameCell {row} /></CardHeader>

<Table class="mb-0" borderless striped size="sm">
  <tbody class="fit">
    <tr>
      <th><T id="Actions.titles.Counter" /></th>
      <td><RunsCell {row} /></td>
    </tr>

    {#if type === 'Timers'}
      <tr>
        <th><T id="Actions.titles.Interval" /></th>
        <td>{rtrim(row.dur.split('.')[0], 's')}{row.dur.endsWith('s') ? 's' : ''}</td>
      </tr>
    {:else if type === 'Schedules'}
      <tr><th><T id="Actions.titles.Schedule" /></th><td>{row.dur}</td></tr>
    {/if}
  </tbody>
</Table>

<style>
  tbody.fit th {
    padding-left: 0.5rem;
    width: 1%;
    white-space: nowrap;
    margin-right: 1rem;
  }

  tbody.fit td {
    text-align: left;
    padding-right: 0.5rem;
  }
</style>
