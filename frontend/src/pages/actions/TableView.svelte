<script lang="ts">
  import Table, { type TableColumn } from 'svelte-table'
  import NameCell from './NameCell.svelte'
  import RunsCell from './RunsCell.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { type TriggerInfo } from '../../api/notifiarrConfig'
  import { val, dur } from './run'

  type Props = { rows: TriggerInfo[] }
  const { rows }: Props = $props()

  const columns: TableColumn<TriggerInfo>[] = [
    {
      title: $_(`Actions.titles.Type`),
      key: 'kind',
      value: (row: TriggerInfo) => $_(`Actions.titles.${row.kind}`),
      sortable: true,
      headerClass: 'ps-2',
      class: 'ps-2',
    },
    {
      title: $_(`Actions.titles.Action`),
      key: 'name',
      renderComponent: NameCell,
      value: val,
      sortable: true,
    },
    {
      title: $_(`Actions.titles.Counter`),
      key: 'runs',
      renderComponent: RunsCell,
      value: (row: TriggerInfo) => row.runs,
      sortable: true,
    },
    {
      title: $_(`Actions.titles.When`),
      key: 'when',
      value: dur,
      sortable: true,
      class: 'pe-2',
    },
  ]
</script>

<Table
  rowKey="name"
  classNameInput="form-control form-control-sm"
  classNameSelect="form-select form-select-sm"
  classNameTable="table table-sm table-striped table-borderless mb-0"
  classNameRow="table-row"
  sortOrders={[-1, 1, 0]}
  sortBy="name"
  sortOrder={1}
  {columns}
  {rows} />
