<script lang="ts">
  import { Card, CardBody, Row, Col, CardHeader } from '@sveltestrap/sveltestrap'
  import T from '../../includes/Translate.svelte'
  import Expvar from './Expvar.svelte'
  import { mapLength } from '../../includes/util'

  type Props = { id: string; data?: Record<string, any> | null }
  const { id, data }: Props = $props()
</script>

<Card class="mt-2" color="secondary" outline>
  <CardHeader>
    <h5><T id={`metrics.${id}.title`} /></h5>
    <small class="text-muted"><T id={`metrics.${id}.description`} /></small>
  </CardHeader>
  {#if data && Object.keys(data).length > 0}
    <Row>
      {#each Object.entries(data).sort((a, b) => mapLength(b) - mapLength(a)) as [source, sub]}
        <Col md={6}><Expvar title={source} data={sub} /></Col>
      {/each}
    </Row>
  {:else}
    <CardBody><T id={`metrics.${id}.empty`} /></CardBody>
  {/if}
</Card>
