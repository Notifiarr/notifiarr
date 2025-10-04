<script lang="ts">
  import { Card, Col, Row as BSRow } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import Action from './ActionCard.svelte'
  import type { TriggerInfo } from '../../api/notifiarrConfig'

  type Props = { type: 'Triggers' | 'Timers' | 'Schedules'; rows: TriggerInfo[] }
  const { type, rows }: Props = $props()
  const color = (runs: number) => (runs > 0 ? 'success-subtle' : 'info-subtle')
</script>

{#if rows.length > 0}
  <BSRow>
    <Col>
      <h4 class="my-2"><T id={`Actions.titles.${type}`} /></h4>
      <T id={`Actions.descriptions.${type}`} />
    </Col>
  </BSRow>

  <BSRow class="mb-3">
    {#each rows as row}
      <Col sm={6} xl={4} xxl={3}>
        <Card class="mt-3" outline color={color(Number(row.runs))}>
          <Action {type} {row} />
        </Card>
      </Col>
    {/each}
  </BSRow>
{/if}
