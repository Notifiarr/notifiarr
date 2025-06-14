<script lang="ts">
  import { Card, CardBody, CardHeader, Table } from '@sveltestrap/sveltestrap'
  import T from '../../includes/Translate.svelte'
  import { formatBytes } from '../../includes/util'

  type Props = { title?: string; id?: string; data?: Record<string, any> | null }
  const { title, id, data = null }: Props = $props()
</script>

{#snippet row(name: string, count: number)}
  <tr>
    <th>{name.includes('Bytes') ? formatBytes(count) : count}</th>
    <td>{name}</td>
  </tr>
{/snippet}

<Card class="mt-3" color="secondary-subtle">
  <CardHeader>
    {#if title}
      {title}
    {:else}
      <h5><T id={`metrics.${id}.title`} /></h5>
      <small class="text-muted"><T id={`metrics.${id}.description`} /></small>
    {/if}
  </CardHeader>

  {#if data && Object.keys(data).length > 0}
    <Table striped responsive class="mb-0">
      <tbody class="fit">
        {#each Object.entries(data) as [name, count]}
          {#if typeof count === 'object'}
            {#each Object.entries(count) as [endpoint, subCount]}
              {@render row(`${name} - ${endpoint}`, Number(subCount))}
            {/each}
          {:else}
            {@render row(name, count)}
          {/if}
        {/each}
      </tbody>
    </Table>
  {:else if id}
    <CardBody><T id={`metrics.${id}.empty`} /></CardBody>
  {/if}
</Card>

<style>
  /* Make the left-side headers fit the max-content length. */
  tbody.fit th {
    white-space: nowrap;
    width: 1%;
    padding-right: 0.2rem;
    text-align: right;
  }
</style>
