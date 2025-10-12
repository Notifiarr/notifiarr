<!-- David Newhall II, May 2025, Notifiarr, LLC. -->
<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    FormGroup,
    Input,
    InputGroup,
    Label,
    type InputType,
  } from '@sveltestrap/sveltestrap'
  import {
    faEye,
    faEyeSlash,
    faArrowUpFromBracket,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import {
    faQuestionCircle,
    faExclamationCircle,
  } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import T, { _ } from './Translate.svelte'
  import { type Snippet } from 'svelte'
  import Fa from './Fa.svelte'
  import { slide } from 'svelte/transition'
  import { deepEqual } from './util'
  import { profile } from '../api/profile.svelte'

  interface Props {
    /** Must be unique. Identifies this component. */
    id: string
    /** The label to display above the input. Must be present in translation if not undefined here. */
    label?: string
    /** The placeholder text to display in the input. */
    placeholder?: string
    /** The description to display below the input. Must be present in translation if not undefined here. */
    description?: string
    /** The type of input. Like `text` or `select`. */
    type?: InputType | 'interval' | 'timeout'
    /** Optional tooltip to bind to input. */
    tooltip?: string
    /** Optional value. Should only be used for binding. */
    value?: any
    /** Optional original value. Used to check for changes.*/
    original?: any
    /** Optional badge to display on the input header. */
    badge?: string
    /** Optional options for select input. */
    options?: Option[] | undefined
    /** Optional validation function. */
    validate?: (id: string, value: any) => string | undefined
    /** Optional input-box prefix attachment. */
    pre?: Snippet
    /** Optional input-box suffix attachment. */
    post?: Snippet
    /** Optional children to render inside the input. Useful for select options. */
    children?: Snippet
    /** Optional message to display below the input. */
    msg?: Snippet
    /** Optional inner value for binding. */
    inner?: any
    /** When type is "timeout" this controls if -1 / disabled is an option. */
    noDisable?: boolean
    /** If this env var is set a notice is provided to the user as such. */
    envVar?: string
    /** Optional other attributes to apply to the input. */
    [key: string]: any
  }

  let {
    id,
    label = $_(`${id}.label`),
    placeholder = $bindable($_(`${id}.placeholder`)),
    description = $_(`${id}.description`),
    type = 'text',
    tooltip = $_(`${id}.tooltip`),
    value = $bindable(undefined),
    original = value,
    options = undefined,
    validate,
    pre,
    children,
    badge = '',
    post,
    msg,
    inner = $bindable(),
    noDisable = false,
    envVar,
    ...rest
  }: Props = $props()

  type Option = { value: string | number | boolean; name: string; disabled?: boolean }

  let showTooltip = $state(false)
  let changed = $derived(original !== null && !deepEqual(value, original))
  let currType = $derived(type)
  let passIcon = $derived(currType === 'password' ? faEyeSlash : faEye)
  let feedback = $state('')
  const inputClass = $derived(!!feedback ? 'is-invalid' : changed ? 'is-valid' : '')
  const env = $derived('DN_' + envVar?.toUpperCase())
  const hasEnv = $derived(!rest.disabled && !!envVar && !!$profile.environment?.[env])

  $effect(() => {
    placeholder = placeholder == id + '.placeholder' ? '' : placeholder
  })

  $effect(() => {
    feedback = validate?.(id, value) ?? ''
  })

  function toggleTooltip(e: Event | undefined = undefined) {
    e?.preventDefault()
    showTooltip = !showTooltip
  }

  function togglePassword(e: Event | undefined = undefined) {
    e?.preventDefault()
    currType = currType === 'password' ? 'text' : 'password'
  }

  if (type === 'interval') {
    currType = 'select'
    options = [
      { value: '0s', name: $_('words.select-option.ChecksDisabled') },
      { value: '1m0s', name: '1 ' + $_('words.select-option.minute') },
      { value: '2m0s', name: '2 ' + $_('words.select-option.minutes') },
      { value: '3m0s', name: '3 ' + $_('words.select-option.minutes') },
      { value: '4m0s', name: '4 ' + $_('words.select-option.minutes') },
      { value: '5m0s', name: '5 ' + $_('words.select-option.minutes') },
      { value: '6m0s', name: '6 ' + $_('words.select-option.minutes') },
      { value: '7m0s', name: '7 ' + $_('words.select-option.minutes') },
      { value: '8m0s', name: '8 ' + $_('words.select-option.minutes') },
      { value: '9m0s', name: '9 ' + $_('words.select-option.minutes') },
      { value: '10m0s', name: '10 ' + $_('words.select-option.minutes') },
      { value: '15m0s', name: '15 ' + $_('words.select-option.minutes') },
      { value: '20m0s', name: '20 ' + $_('words.select-option.minutes') },
      { value: '25m0s', name: '25 ' + $_('words.select-option.minutes') },
      { value: '30m0s', name: '30 ' + $_('words.select-option.minutes') },
    ]
  }

  if (type === 'timeout') {
    currType = 'select'
    options = [
      { value: '0s', name: $_('words.select-option.NoTimeout') },
      { value: '1s', name: '1 ' + $_('words.select-option.seconds') },
      { value: '2s', name: '2 ' + $_('words.select-option.seconds') },
      { value: '3s', name: '3 ' + $_('words.select-option.seconds') },
      { value: '4s', name: '4 ' + $_('words.select-option.seconds') },
      { value: '5s', name: '5 ' + $_('words.select-option.seconds') },
      { value: '10s', name: '10 ' + $_('words.select-option.seconds') },
      { value: '15s', name: '15 ' + $_('words.select-option.seconds') },
      { value: '30s', name: '30 ' + $_('words.select-option.seconds') },
      { value: '1m0s', name: '1 ' + $_('words.select-option.minute') },
      { value: '2m0s', name: '2 ' + $_('words.select-option.minutes') },
      { value: '3m0s', name: '3 ' + $_('words.select-option.minutes') },
      { value: '4m0s', name: '4 ' + $_('words.select-option.minutes') },
      { value: '5m0s', name: '5 ' + $_('words.select-option.minutes') },
      { value: '6m0s', name: '6 ' + $_('words.select-option.minutes') },
      { value: '7m0s', name: '7 ' + $_('words.select-option.minutes') },
      { value: '8m0s', name: '8 ' + $_('words.select-option.minutes') },
      { value: '9m0s', name: '9 ' + $_('words.select-option.minutes') },
      { value: '10m0s', name: '10 ' + $_('words.select-option.minutes') },
    ]
    if (!noDisable)
      options.unshift({ value: '-1s', name: $_('words.select-option.InstanceDisabled') })
  }
</script>

<div class="input">
  <FormGroup>
    <Label for={id}>
      {@html label}
      {#if badge}
        <Badge color="secondary" style="margin-left: 0.5rem;">{badge}</Badge>
      {/if}
    </Label>
    <InputGroup>
      {#if tooltip != id + '.tooltip' || (envVar && !rest.disabled)}
        <Button
          color="secondary"
          onclick={toggleTooltip}
          outline
          style="width:44px;"
          title={$_('phrases.ShowMore')}>
          {#if showTooltip}
            <Fa
              i={faArrowUpFromBracket}
              c1="dimgray"
              d1="gainsboro"
              c2="orange"
              scale="1.5x" />
          {:else}
            <Fa
              i={hasEnv ? faExclamationCircle : faQuestionCircle}
              c1="dimgray"
              d1="gainsboro"
              c2={hasEnv ? 'red' : 'orange'}
              d2={hasEnv ? 'mediumvioletred' : 'orange'}
              scale="1.5x" />
          {/if}
        </Button>
      {/if}
      {@render pre?.()}
      <Input
        {id}
        class="{inputClass} {changed ? 'changed' : ''}"
        type={currType as InputType}
        bind:inner
        bind:value
        bind:checked={value}
        autocomplete="off"
        {placeholder}
        {...rest}>
        {#if children}
          {@render children()}
        {:else if options}
          <!-- render provided options. -->
          {#if !options.map(o => o.value).includes(value)}
            <!-- If the current value is not in the options list, add it. -->
            <option {value} selected>
              {value} ({$_('words.select-option.custom')})
            </option>
          {/if}
          <!-- Create a select option list from `options` input. -->
          {#each options as o}
            <option value={o.value} selected={value === o.value} disabled={o.disabled}>
              {o.name}
            </option>
          {/each}
        {:else if typeof value === 'boolean' && type === 'select'}
          <!-- Create a boolean select-option list: Enabled/Disabled
           If the name of the input ends with 'disabled', then the values are inverted.
           -->
          <option
            value={id.endsWith('disabled') ? true : false}
            selected={value === id.endsWith('disabled') ? true : false}>
            {$_('words.select-option.Disabled')}
          </option>
          <option
            value={id.endsWith('disabled') ? false : true}
            selected={value === id.endsWith('disabled') ? false : true}>
            {$_('words.select-option.Enabled')}
          </option>
        {/if}
      </Input>

      <!-- Include a password visibility toggler. -->
      {#if type === 'password'}
        <Button
          type="button"
          outline
          onclick={togglePassword}
          style="width:44px;"
          title="Toggle password visibility">
          <Fa
            i={passIcon}
            c1="royalblue"
            c2="orange"
            d1="orange"
            d2="dodgerblue"
            scale="1.5x" />
        </Button>
      {/if}
      {@render post?.()}
    </InputGroup>
    <div class="text-danger">{feedback}</div>

    {#if showTooltip}
      <div transition:slide>
        <Card body class="mt-1" color="warning" outline>
          {#if !rest.disabled}
            <ul class="mb-0">
              <li><T id="phrases.EnvironmentVariable" variableName={env} /></li>
            </ul>
            {#if hasEnv}
              <p class="mt-2 mb-0"><T id="phrases.VariableDescription" /></p>
            {/if}
          {/if}
          {#if tooltip != id + '.tooltip'}<p class="mt-2 mb-0">{@html tooltip}</p>{/if}
        </Card>
      </div>
    {/if}

    {#if description}<small class="text-muted">{@html description}</small>{/if}
    {@render msg?.()}
  </FormGroup>
</div>

<style>
  .input {
    margin-bottom: 1rem;
  }

  /** Allows textarea to be resized vertically on mobile. */
  .input :global(textarea) {
    resize: vertical;
  }

  .input :global(label) {
    font-weight: 550;
    font-family: Verdana, Geneva, Tahoma, sans-serif;
  }

  .input :global(.changed) {
    background-color: rgba(205, 92, 92, 0.322);
  }
</style>
