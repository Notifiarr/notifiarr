<!-- Test every configured instance for reachability. -->
<script lang="ts" module>
  import { CardBody, CardTitle, Card, Table, CardHeader } from '@sveltestrap/sveltestrap'
  import { getUi } from '../../api/fetch'
  import type { CheckAllOutput, TestResult } from '../../api/notifiarrConfig'
  import Nodal from '../../includes/Nodal.svelte'
  import { faTents } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { age } from '../../includes/util'
  import { profile } from '../../api/profile.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import type { TableColumn } from 'svelte-table'
  import SvelteTable from 'svelte-table'
  import NameCell, { type Result } from './NameCell.svelte'

  export const page = {
    type: 'modal' as const,
    id: 'TestAllInstances',
    i: faTents,
    c1: 'coral',
    c2: 'steelblue',
    d1: 'wheat',
    d2: 'lime',
  }
</script>

<script lang="ts">
  let isOpen = $state(false)
  const get = async () => await getUi('checkAllInstances', true, 16000)
  export const toggle = () => (isOpen = !isOpen)
  let expanded = $state<number[]>([])

  const rows = (data: CheckAllOutput): Result[] => {
    let idx = 0
    return Object.keys(data)
      .filter(key => typeof data[key as keyof CheckAllOutput] !== 'number')
      .map(app => {
        return (data[app as keyof CheckAllOutput] as TestResult[])?.map(
          (item, id) => ({ ...item, app, idx: idx++, id }) as Result,
        )
      })
      .flat()
  }

  const expand = (idx: number) => {
    if (expanded.includes(idx)) expanded.splice(expanded.indexOf(idx), 1)
    else expanded.push(idx)
  }

  const color = (row: Result): string => {
    if (row.status === 200) return 'success'
    else if (row.config.timeout == '-1s') return 'warning'
    else return 'danger'
  }

  const columns: TableColumn<Result>[] = [
    {
      class: row => 'fit text-' + color(row),
      headerClass: 'fit',
      title: $_(`TestAllInstances.Instance`),
      key: 'instance',
      value: row => `${row.app} ${row.id! + 1}`,
      renderComponent: { component: NameCell, props: { expand } },
      sortable: true,
    },
    {
      title: $_(`TestAllInstances.Message`),
      key: 'message',
      value: (row: TestResult) => row.message,
      sortable: true,
    },
    {
      class: 'fit text-muted',
      headerClass: 'fit',
      title: $_(`Integrations.starrTitles.Elapsed`),
      key: 'elapsed',
      value: (row: TestResult) => row.elapsed,
      sortable: true,
    },
  ]
</script>

<Nodal {get} bind:isOpen title={page.id + '.title'} fa={page} size="xl" full>
  {#snippet children(resp)}
    {#if resp?.ok}
      <ul class="mb-1">
        <li>
          <T id="TestAllInstances.Concurrency" workerCount={resp.body.workers} />
        </li>
        <li>
          <T id="TestAllInstances.Tested" testCount={resp.body.instances} />
        </li>
        <li>
          <T
            id="TestAllInstances.TestElapsed"
            timeDuration={age(resp.body.elapsed, true)} />
        </li>
        <li>
          <T
            id="TestAllInstances.Updated"
            timeDuration={age(profile.now - resp.body.timeMS)} />
        </li>
        <li><T id="TestAllInstances.clickRefresh" /></li>
      </ul>
      <div class="svtable">
        <SvelteTable
          rowKey="idx"
          bind:expanded
          classNameInput="form-control form-control-sm"
          classNameSelect="form-select form-select-sm"
          classNameTable="table table-sm mb-0 pb-0"
          sortOrders={[-1, 1, 0]}
          {columns}
          rows={rows(resp.body)}>
          <div slot="expanded" let:row>
            {@const res = row as Result}
            <Card outline color={color(res)}>
              <CardHeader><CardTitle>{res.config.name}</CardTitle></CardHeader>
              <Table borderless striped responsive size="sm" class="mb-0">
                <tbody>
                  <tr>
                    <td class="fit2"><T id="words.instance-options.timeout.label" /></td>
                    <td>
                      {res.config.timeout}
                      <span class="text-muted">
                        {#if res.config.timeout == '0s'}
                          ({$_(`words.select-option.NoTimeout`)})
                        {:else if res.config.timeout == '-1s'}
                          ({$_(`words.select-option.InstanceDisabled`)})
                        {/if}
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td class="fit2"><T id="words.instance-options.interval.label" /></td>
                    <td>{res.config.interval}</td>
                  </tr>
                  <tr>
                    <td class="fit2"><T id="words.instance-options.validSsl.label" /></td>
                    <td>{res.config.validSsl}</td>
                  </tr>
                  <tr>
                    <td class="fit2"><T id="words.instance-options.deletes.label" /></td>
                    <td>{res.config.deletes}</td>
                  </tr>
                  <tr>
                    <td class="fit2"><T id="Integrations.starrTitles.Elapsed" /></td>
                    <td>{res.elapsed}</td>
                  </tr>
                  <tr>
                    <td class="fit2"><T id="TestAllInstances.Status" /></td>
                    <td>{res.status === 200 ? '✅' : '❌'} {res.status}</td>
                  </tr>
                </tbody>
              </Table>
            </Card>
          </div>
        </SvelteTable>
      </div>
    {:else}
      <Card color="danger" outline>
        <CardHeader>
          <CardTitle class="text-danger"><T id="phrases.ERROR" /></CardTitle>
        </CardHeader>
        <CardBody>{resp?.body}</CardBody>
      </Card>
    {/if}
  {/snippet}
  {#snippet footer(resp)}
    <T id="TestAllInstances.footer" />
  {/snippet}
</Nodal>

<style>
  .svtable :global(.fit) {
    white-space: nowrap;
    width: 1%;
  }

  .fit2 {
    white-space: nowrap;
    width: 1%;
    padding-left: 0.5rem;
  }

  .svtable :global(tbody tr:last-child th),
  .svtable :global(tbody tr:last-child td) {
    border-bottom: none;
  }
</style>
