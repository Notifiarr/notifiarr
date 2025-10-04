<script lang="ts" module>
  import { get } from 'svelte/store'
  import T, { _ } from '../../includes/Translate.svelte'
  /** Call this from the validator method you pass in to the module to validate its values. */
  export const validator = (id: string, value: any): string => {
    if (id.endsWith('daysOfWeek') && !value)
      return get(_)('scheduler.required.daysOfWeek')
    if (id.endsWith('daysOfMonth') && !value)
      return get(_)('scheduler.required.daysOfMonth')
    if (id.endsWith('atTimes') && (!value || value.length === 0))
      return get(_)('scheduler.required.atTimes')
    return ''
  }
</script>

<script lang="ts">
  import {
    Button,
    Col,
    FormGroup,
    Input,
    InputGroup,
    Label,
    Row,
  } from '@sveltestrap/sveltestrap'
  import { Frequency, type CronJob } from '../../api/notifiarrConfig'
  import { deepCopy, deepEqual } from '../../includes/util'
  import Select from 'svelte-select'
  import Fa from '../../includes/Fa.svelte'
  import { faRightToLine } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { cronDesc, cronTimes, weekdays } from './schedule'

  type Props = {
    cron: CronJob
    original: CronJob
    validate?: (id: string, value: any) => string | undefined
  }
  let { cron = $bindable(), original, validate }: Props = $props()

  let feedback = $state<Record<string, string>>({})

  /** The new time to add. Part of the time picker.*/
  let newTime = $state<number[]>([0, 0, 0])
  /** Add a new time to the list of times. Part of the time picker.*/
  const addNewTime = () => {
    if (!cron.atTimes) cron.atTimes = []
    if (cron.atTimes.some(t => deepEqual(t, newTime))) return
    cron.atTimes.push(newTime)
    cron.atTimes = cron.atTimes.sort((a, b) => a[0] - b[0] || a[1] - b[1] || a[2] - b[2])
    newTime = [0, 0, 0]
  }

  /** Update form values when the frequency changes. */
  const newFrequency = () => {
    /** Put everything back if the frequency is set back to the original. */
    if (cron.frequency === original.frequency) {
      cron = deepCopy(original)
    } /** Reset everything when the frequency changes. */ else {
      cron.atTimes = []
      cron.daysOfWeek = []
      cron.daysOfMonth = []
    }
    validateDays(0)
    feedback['atTimes'] = validate?.('cron.atTimes', cron.atTimes?.length ?? 0) ?? ''
  }

  let dayVal = Object.entries(weekdays).map(([v, label]) => ({ label, value: Number(v) }))

  const validateDays = (subtract: number) => {
    if (cron.frequency === Frequency.Weekly) {
      feedback['daysOfWeek'] =
        validate?.('cron.daysOfWeek', (cron.daysOfWeek?.length ?? 0) - subtract) ?? ''
      feedback['daysOfMonth'] = validate?.('cron.daysOfMonth', 1) ?? ''
    } else if (cron.frequency === Frequency.Monthly) {
      feedback['daysOfMonth'] =
        validate?.('cron.daysOfMonth', (cron.daysOfMonth?.length ?? 0) - subtract) ?? ''
      feedback['daysOfWeek'] = validate?.('cron.daysOfWeek', 1) ?? ''
    }
  }

  const clear = (e: any) => {
    let cleared = e.detail
    if (!Array.isArray(cleared)) cleared = [cleared]
    validateDays(cleared.length)
    sortDays()
  }

  export const reset = () => {
    feedback = {}
  }

  /** Sort the days every time the form changes. */
  const sortDays = () => {
    // Sort the values displayed in text.
    cron.daysOfWeek = cron.daysOfWeek?.sort((a, b) => a - b)
    cron.daysOfMonth = cron.daysOfMonth?.sort((a, b) => a - b)
  }

  /** Remove a time from the list of times. */
  const deleteTime = (e: CustomEvent<any>) => {
    let value = e.detail
    if (!Array.isArray(value)) value = [e.detail]
    for (let v of value) {
      v = v.value
      if (v.length === 2) v = '0:0:' + v
      if (v.length === 5) v = '0:' + v
      cron.atTimes = cron.atTimes
        ?.filter(t => !deepEqual(t, v.split(':').map(Number)))
        .sort((a, b) => a[0] - b[0] || a[1] - b[1] || a[2] - b[2])
    }
  }
</script>

<div class="cron-scheduler mb-2">
  <Row>
    <Col>
      <Label><T id="scheduler.title" /></Label>
      <FormGroup floating label={$_('scheduler.frequency')} spacing="mb-1">
        <Input
          type="select"
          bind:value={cron.frequency}
          onchange={newFrequency}
          class={cron.frequency === original.frequency ? '' : 'changed is-valid'}>
          <option value={Frequency.DeadCron}><T id="scheduler.ops.noSchedule" /></option>
          <option value={Frequency.Minutely}><T id="scheduler.ops.minutely" /></option>
          <option value={Frequency.Hourly}><T id="scheduler.ops.hourly" /></option>
          <option value={Frequency.Daily}><T id="scheduler.ops.daily" /></option>
          <option value={Frequency.Weekly}><T id="scheduler.ops.weekly" /></option>
          <option value={Frequency.Monthly}><T id="scheduler.ops.monthly" /></option>
        </Input>
        <small class="text-muted">{cronDesc(cron)}</small>
      </FormGroup>
    </Col>
  </Row>

  {#snippet timeInput(idx: number, max: number, disabled: boolean)}
    <Input type="select" bind:value={newTime[idx]} min={0} {max} {disabled} class="tp">
      {#each Array.from({ length: max }, (_, i) => i) as i}
        <option value={i}>{i.toString().padStart(2, '0')}</option>
      {/each}
    </Input>
  {/snippet}

  {#if cron.frequency !== Frequency.DeadCron}
    <table style="width: 100%" class="my-1">
      <tbody class="fit">
        <tr>
          <th>
            <!-- Time picker -->
            <InputGroup>
              <!-- hours -->
              {@render timeInput(
                0,
                24,
                [Frequency.Minutely, Frequency.Hourly].includes(cron.frequency),
              )}
              <!-- minutes -->
              {@render timeInput(1, 60, cron.frequency === Frequency.Minutely)}
              <!-- seconds -->
              {@render timeInput(2, 60, false)}
              <!-- Add button -->
              <Button class="addButton" color="secondary" outline onclick={addNewTime}>
                <Fa
                  i={faRightToLine}
                  scale="1.5"
                  c2="coral"
                  d2="pink"
                  c1="green"
                  d1="seagreen" />
              </Button>
            </InputGroup>
          </th>
          <td>
            <!-- Input box with times. Used to delete them (by clicking the x). -->
            <div class="time-container p-0">
              <Select
                on:input={() =>
                  (feedback['atTimes'] =
                    validate?.('cron.atTimes', cron.atTimes?.length ?? 0) ?? '')}
                class="multiselect {cron.atTimes?.length &&
                deepEqual(cron.atTimes, original.atTimes)
                  ? ''
                  : 'changed ' + (cron.atTimes?.length ? 'is-valid' : 'is-invalid')}"
                multiple
                hideEmptyState
                clearable={false}
                multiFullItemClearable={false}
                inputAttributes={{ readonly: true }}
                placeholder="⬅ {$_('scheduler.AddATime')}"
                on:clear={deleteTime}
                value={cronTimes(cron)} />
            </div>
          </td>
        </tr>
        <tr>
          <td colspan="2">
            <span class="text-danger">{feedback['atTimes']}</span>
          </td>
        </tr>
      </tbody>
    </table>
  {/if}

  {#if cron.frequency === Frequency.Weekly}
    <Select
      on:clear={clear}
      on:change={() => (sortDays(), validateDays(0))}
      class="form-control multiselect {cron.daysOfWeek?.length &&
      deepEqual(cron.daysOfWeek, original.daysOfWeek)
        ? ''
        : 'changed ' + (cron.daysOfWeek?.length ? 'is-valid' : 'is-invalid')}"
      placeholder={$_('scheduler.daysOfWeek')}
      bind:justValue={cron.daysOfWeek}
      value={cron.daysOfWeek?.map(d => ({ value: d, label: weekdays()[d] })) ?? undefined}
      multiple
      searchable
      clearable
      items={dayVal}>
      <div slot="empty"><p class="text-center mt-3">{$_('scheduler.empty')}</p></div>
    </Select>
    <span class="text-danger">{feedback['daysOfWeek']}</span>
  {:else if cron.frequency === Frequency.Monthly}
    <Select
      on:clear={clear}
      on:change={() => (sortDays(), validateDays(0))}
      class="form-control multiselect {cron.daysOfMonth?.length &&
      deepEqual(cron.daysOfMonth, original.daysOfMonth)
        ? ''
        : 'changed ' + (cron.daysOfMonth?.length ? 'is-valid' : 'is-invalid')}"
      placeholder={$_('scheduler.daysOfMonth')}
      bind:justValue={cron.daysOfMonth}
      value={cron.daysOfMonth?.map(d => ({ value: d, label: `${d}` })) ?? undefined}
      multiple
      searchable
      clearable
      placeholderAlwaysShow={true}
      items={Array.from({ length: 31 }, (_, i) => i + 1)}>
      <div slot="empty"><p class="text-center mt-3">{$_('scheduler.empty')}</p></div>
    </Select>
    <span class="text-danger">{feedback['daysOfMonth']}</span>
  {/if}
</div>

<style>
  /* Make the stuff match other pages. */
  .cron-scheduler {
    font-family: Verdana, Geneva, Tahoma, sans-serif;
  }

  .cron-scheduler :global(label) {
    font-weight: 550;
  }

  /** Make the timepicker look correct. (smash the inputs together) */
  .cron-scheduler :global(.input-group),
  .cron-scheduler :global(.tp) {
    display: inline-table !important;
    width: auto;
    margin: 0 !important;
    height: 100% !important;
    max-height: 110px !important;
  }

  .cron-scheduler :global(.tp) {
    width: 34px !important;
    padding: 0 !important;
    text-indent: 6px !important;
    background: none !important;
  }

  /** remove the shadow that looks weird on this box.*/
  .time-container :global(.multiselect) {
    --list-shadow: 0 !important;
    --list-border: 0 !important;
  }

  /* Make left hand table column "fit" content. */
  tbody.fit th {
    white-space: nowrap;
    width: 1%;
    padding-right: 5px;
    height: 45px !important;
  }

  .fit :global(.addButton) {
    height: 100%;
    margin-bottom: 3px;
    max-height: 110px !important;
  }

  .cron-scheduler :global(.changed) {
    background-color: rgba(205, 92, 92, 0.322) !important;
  }

  .cron-scheduler :global(.is-invalid) {
    --border: var(--bs-border-width) solid var(--bs-form-invalid-border-color);
    --border-hover: var(--bs-border-width) solid var(--bs-form-invalid-border-color);
  }
  .cron-scheduler :global(.is-valid) {
    --border: var(--bs-border-width) solid var(--bs-form-valid-border-color);
    --border-hover: var(--bs-border-width) solid var(--bs-form-valid-border-color);
  }
</style>
