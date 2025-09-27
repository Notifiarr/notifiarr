<script lang="ts">
  import { Card, Col, Row } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import { type TriggerInfo } from '../../api/notifiarrConfig'
  import Action from './Action.svelte'

  type Obj = Record<string, TriggerInfo>
  type Props = { type: 'Triggers' | 'Timers' | 'Schedules'; obj: Obj }
  const { type, obj }: Props = $props()
  const color = (runs: number) => (runs > 0 ? 'success-subtle' : 'info-subtle')
</script>

<Row>
  <Col>
    <h4><T id={`Actions.titles.${type}`} /></h4>
    <T id={`Actions.descriptions.${type}`} />
  </Col>
</Row>

<Row>
  {#each Object.entries(obj ?? {}) as [key, info]}
    <Col sm={6} xl={4} xxl={2}>
      <Card class="mt-3" outline color={color(Number(info.runs))}>
        <Action {type} {info} />
      </Card>
    </Col>
  {/each}
</Row>
