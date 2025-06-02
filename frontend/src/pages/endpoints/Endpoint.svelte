<script lang="ts">
  import Input from '../../includes/Input.svelte'

  import { Col, Input as Box, Row, Button } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import { Frequency, type Endpoint } from '../../api/notifiarrConfig'
  import type { ChildProps } from '../../includes/Instances.svelte'
  import CheckedInput from '../../includes/CheckedInput.svelte'

  let {
    form = $bindable(),
    original,
    app,
    index = 0, // unused but matches our ChildProps interface.
    validate,
  }: ChildProps<Endpoint> = $props()
</script>

<!-- query?: Record<string, null | string[]>;
    header?: Record<string, null | string[]>;

   * Frequency to configure the job. Pass 0 disable the cron.
   */
  frequency: number;
  /**
   * Interval for Daily, Weekly and Monthly Frequencies. 1 = every day/week/month, 2 = every other, and so on.
   */
  interval: number;
  /**
   * AtTimes is a list of 'hours, minutes, seconds' to schedule for Daily/Weekly/Monthly frequencies.
   * Also used in Minutely and Hourly schedules, a bit awkwardly.
   */
  atTimes?: number[][];
  /**
   * DaysOfWeek is a list of days to schedule. 0-6. 0 = Sunday.
   */
  daysOfWeek?: Weekday[];
  /**
   * DaysOfMonth is a list of days to schedule. 1 to 31 or -31 to -1 to count backward.
   */
  daysOfMonth?: number[];
  /**
   * Months to schedule. 1 to 12. 1 = January.
   */
  months?: number[];

     -->
<div class="endpoint">
  <Row>
    <Col md={6}>
      <Input
        id={app.id + '.name'}
        bind:value={form.name}
        original={original?.name}
        {validate} />
    </Col>
    <Col md={6}>
      <Input
        id={app.id + '.template'}
        bind:value={form.template}
        original={original?.template}
        {validate}>
        {#snippet post()}
          <Button color="secondary" outline>
            <Box type="checkbox" id={app.id + '.follow'} bind:checked={form.follow} />
          </Button>
        {/snippet}
      </Input>
    </Col>
    <Col md={2}>
      <Input
        type="select"
        id={app.id + '.method'}
        bind:value={form.method}
        original={original?.method}
        {validate}
        options={['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'].map(m => ({
          name: m,
          value: m,
        }))} />
    </Col>
    <Col md={10}>
      <CheckedInput id="url" {app} {index} bind:form bind:original {validate} />
    </Col>
    <Col lg={12}>
      <Input
        rows={2}
        type="textarea"
        id={app.id + '.body'}
        bind:value={form.body}
        original={original?.body}
        {validate} />
    </Col>
    <Col md={3}>
      <Input
        type="select"
        id={app.id + '.frequency'}
        bind:value={form.frequency}
        original={original?.frequency}
        {validate}
        options={[
          { name: 'Unscheduled', value: Frequency.DeadCron },
          { name: 'Every Minute', value: Frequency.Minutely },
          { name: 'By the Hour', value: Frequency.Hourly },
          { name: 'By the Day', value: Frequency.Daily },
          { name: 'By the Week', value: Frequency.Weekly },
          { name: 'By the Month', value: Frequency.Monthly },
        ]} />
    </Col>
  </Row>
</div>

<style>
  .endpoint :global(.changed) {
    background-color: rgba(205, 92, 92, 0.322) !important;
  }
</style>
