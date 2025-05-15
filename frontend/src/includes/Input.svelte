<!-- David Newhall II, May 2025, Notifiarr, LLC. -->
<script lang="ts">
  import {
    Button,
    Card,
    FormGroup,
    Input,
    InputGroup,
    Label,
    type InputType,
  } from '@sveltestrap/sveltestrap'
  import { _ } from './Translate.svelte'
  import type { SvelteComponent } from 'svelte'
  import {
    faCircleDot,
    faEye,
    faEyeSlash,
    faQuestionCircle,
  } from '@fortawesome/free-solid-svg-icons'
  import Fa from 'svelte-fa'

  /** Must be unique. Identifies this component. */
  export let id: string
  /** The name of the input. Defaults to the id. Pass undefined to not include a name. */
  export let name: string = id
  /** The label to display above the input. Must be present in translation if not undefined here. */
  export let label: string | undefined = $_(`${id}.label`)
  /** The placeholder text to display in the input. */
  export let placeholder: string | undefined = $_(`${id}.placeholder`)
  /** The description to display below the input. Must be present in translation if not undefined here. */
  export let description: string | undefined = $_(`${id}.description`)
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

  type Option = { value: string | number | boolean; name: string; disabled?: boolean }

  let input: SvelteComponent
  let showTooltip = false

  $: icon = showTooltip ? faCircleDot : faQuestionCircle
  $: currType = type
  $: passIcon = currType === 'password' ? faEyeSlash : faEye
  $: iconClass = showTooltip ? 'text-danger' : 'text-secondary'
  $: placeholder = placeholder == id + '.placeholder' ? undefined : placeholder

  function toggleTooltip(e: Event | undefined = undefined) {
    e?.preventDefault()
    showTooltip = !showTooltip
  }

  function togglePassword(e: Event | undefined = undefined) {
    e?.preventDefault()
    currType = currType === 'password' ? 'text' : 'password'
  }
</script>

<div class="input">
  <FormGroup>
    <Label for={id}>{@html label}</Label>
    <InputGroup>
      {#if tooltip != id + '.tooltip'}
        <Button color="warning" on:click={toggleTooltip} outline>
          <Fa {icon} class={iconClass} />
        </Button>
      {/if}
      <slot name="pre" />
      <Input
        {id}
        {name}
        type={currType}
        bind:this={input}
        bind:value
        bind:checked={value}
        autocomplete={name.includes('noauto') ? 'off' : undefined}
        {placeholder}
        {readonly}
        {disabled}
        {rows}
        {min}
        {max}>
        <!-- Create a boolean select option list. -->
        {#if typeof value === 'boolean' && type === 'select'}
          <option value={false} selected={value === false}>
            {$_('words.select-option.Disabled')}
          </option>
          <option value={true} selected={value === true}>
            {$_('words.select-option.Enabled')}
          </option>
        {/if}
        <!-- If the current value is not in the options list, add it. -->
        {#if options}
          {#if !options.map(o => o.value).includes(value)}
            <option {value} selected>
              {value} ({$_('words.select-option.custom')})
            </option>
          {/if}
          <!-- Create a select option list from input. -->
          {#each options as o}
            <option value={o.value} selected={value === o.value} disabled={o.disabled}>
              {o.name}
            </option>
          {/each}
        {/if}
        <slot />
      </Input>
      <!-- Including a password visibility toggler. -->
      {#if type === 'password'}
        <Button type="button" outline on:click={togglePassword}>
          <Fa icon={passIcon} class="text-warning" />
        </Button>
      {/if}
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
