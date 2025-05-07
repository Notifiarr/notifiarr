<script lang="ts">
  import {
    Button,
    Card,
    FormGroup,
    Icon,
    Input,
    InputGroup,
    Label,
    type InputType,
  } from '@sveltestrap/sveltestrap'
  import { _ } from './Translate.svelte'
  import type { SvelteComponent } from 'svelte'

  /** Must be unique. Identifies this component. */
  export let id: string
  /** The label to display above the input. */
  export let label: string = $_(`${id}.label`)
  /** The placeholder text to display in the input. */
  export let placeholder: string = $_(`${id}.placeholder`)
  /** The description to display below the input. */
  export let description: string = $_(`${id}.description`)
  /** The type of input. Like `text` or `select`. */
  export let type: InputType = 'text'
  /** Used if you do not want this value changed directly. */
  export let readonly = false
  /** Similar to readonly, but the input dims/greys out. */
  export let disabled = false
  /** Optional tooltip to bind to input. */
  export let tooltip: string = $_(`${id}.tooltip`)
  /** Optional value. Should only be used for binding. */
  export let value: any = undefined
  /** Optional rows for textarea. */
  export let rows: number = 1
  /** Optional min value for number input. */
  export let min: number | undefined = undefined
  /** Optional max value for number input. */
  export let max: number | undefined = undefined
  /** Optional options for select input. */
  export let options: Option[] | undefined = undefined

  type Option = { value: string | number | boolean; name: string }

  let input: SvelteComponent

  $: tooltip = $_(`${id}.tooltip`)
  $: label = $_(`${id}.label`)
  $: placeholder = $_(`${id}.placeholder`)
  $: description = $_(`${id}.description`)

  $: icon = showTooltip ? 'dash-circle' : 'question-circle'
  $: iconClass = showTooltip ? 'text-danger' : 'text-secondary'

  let showTooltip = false
  function toggleTooltip(e: Event | undefined = undefined) {
    e?.preventDefault()
    showTooltip = !showTooltip
  }
</script>

<div class="input">
  <FormGroup>
    <Label for={id}>{@html label}</Label>
    <InputGroup>
      {#if tooltip != id + '.tooltip'}
        <Button color="warning" on:click={toggleTooltip} outline>
          <Icon class={iconClass} name={icon} />
        </Button>
      {/if}
      <slot name="pre" />
      <Input
        {id}
        {type}
        bind:this={input}
        bind:value
        bind:checked={value}
        {placeholder}
        {readonly}
        {disabled}
        {rows}
        {min}
        {max}>
        {#if typeof value === 'boolean' && type === 'select'}
          <option value={false} selected={value === false}>
            {$_('words.select-option.Disabled')}
          </option>
          <option value={true} selected={value === true}>
            {$_('words.select-option.Enabled')}
          </option>
        {/if}
        {#if options}
          {#if !options.map(o => o.value).includes(value)}
            <option {value} selected>
              {value} ({$_('words.select-option.custom')})
            </option>
          {/if}
          {#each options as option}
            <option value={option.value} selected={value === option.value}>
              {option.name}
            </option>
          {/each}
        {/if}
        <slot />
      </Input>
      <slot name="post" />
    </InputGroup>

    {#if showTooltip}
      <Card body class="mt-1" color="warning" outline>
        <p class="mb-0">{@html tooltip}</p>
      </Card>
    {/if}

    {#if description}
      <small class="text-muted">{@html description}</small>
    {/if}
  </FormGroup>
</div>

<style>
  .input {
    margin-bottom: 1rem;
  }
</style>
