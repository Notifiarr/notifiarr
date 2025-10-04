<script lang="ts">
  import Table, { type TableColumn } from 'svelte-table'
  import NameCell from './NameCell.svelte'
  import RunsCell from './RunsCell.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { type Row, val, dur } from './run'

  type Props = { rows: Row[]; filter?: string }
  const { rows, filter = '' }: Props = $props()

  const columns: TableColumn<Row>[] = [
    {
      title: $_(`Actions.titles.Type`),
      key: 'type',
      value: (row: Row) => $_(`Actions.titles.${row.type}`),
      sortable: true,
    },
    {
      title: $_(`Actions.titles.Name`),
      key: 'name',
      renderComponent: NameCell,
      value: val,
      sortable: true,
    },
    {
      title: $_(`Actions.titles.Counter`),
      key: 'runs',
      renderComponent: RunsCell,
      value: (row: Row) => Number(row.runs),
      sortable: true,
    },
    { title: $_(`Actions.titles.When`), key: 'when', value: dur, sortable: true },
  ]

  // Smash all the triggers, timers, and schedules into one array and add a type column.
  const filtered = $derived(
    filter
      ? rows.filter(row => val(row).toLowerCase().includes(filter.toLowerCase()))
      : rows,
  )
</script>

<Table
  rowKey="name"
  classNameInput="form-control form-control-sm"
  classNameSelect="form-select form-select-sm"
  classNameTable="table table-sm table-striped"
  classNameRow="table-row"
  sortOrders={[-1, 1, 0]}
  {columns}
  rows={filtered} />
