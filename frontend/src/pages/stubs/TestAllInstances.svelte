<!-- Test every configured instance for reachability. -->
<script lang="ts" module>
  import { CardBody, CardTitle, Table, Card, CardHeader } from '@sveltestrap/sveltestrap'
  import { getUi } from '../../api/fetch'
  import type { TestResult } from '../../api/notifiarrConfig'
  import Nodal from '../../includes/Nodal.svelte'
  import { faTents } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { age } from '../../includes/util'
  import { profile } from '../../api/profile.svelte'
  import T from '../../includes/Translate.svelte'

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
  const get = async () => await getUi('checkAllInstances', true)
  export const toggle = () => (isOpen = !isOpen)
</script>

<Nodal {get} bind:isOpen title={page.id + '.title'} fa={page} size="xl" full>
  {#snippet children(resp)}
    {#if resp?.ok}
      Updated: {age(profile.now - resp.body.timeMS)}
      <Table class="mb-0" size="sm">
        <thead>
          <tr>
            <th><T id="TestAllInstances.Instance" /></th>
            <th><T id="TestAllInstances.Message" /></th>
          </tr>
        </thead>
        <tbody class="fit">
          {#each Object.keys(resp.body).filter(key => key !== 'timeMS') as key}
            {#each resp.body[key] as TestResult[] as item, idx}
              {@const color = item.status === 200 ? 'success' : 'danger'}
              <tr>
                <th class="text-{color}">{key} {idx + 1}</th>
                <td>{item.message}</td>
              </tr>
            {/each}
          {/each}
        </tbody>
      </Table>
    {:else}
      <Card color="danger" outline>
        <CardHeader>
          <CardTitle class="text-danger"><T id="phrases.ERROR" /></CardTitle>
        </CardHeader>
        <CardBody>{resp?.body}</CardBody>
      </Card>
    {/if}
  {/snippet}
</Nodal>

<style>
  tbody.fit th {
    white-space: nowrap;
    width: 1%;
  }
</style>
