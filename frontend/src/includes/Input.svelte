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
  import {
    faEye,
    faEyeSlash,
    faArrowUpFromBracket,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { faQuestionCircle } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { _ } from './Translate.svelte'
  import type { SvelteComponent, Snippet } from 'svelte'
  import Fa from './Fa.svelte'
  import { slide } from 'svelte/transition'

  interface Props {
    /** Must be unique. Identifies this component. */
    id: string
    /** The name of the input. Defaults to the id. Pass undefined to not include a name. */
    name?: string
    /** The label to display above the input. Must be present in translation if not undefined here. */
    label?: string
    /** The placeholder text to display in the input. */
    placeholder?: string
    /** The description to display below the input. Must be present in translation if not undefined here. */
    description?: string
    /** The type of input. Like `text` or `select`. */
    type?: InputType
    /** Used if you do not want this value changed directly. */
    readonly?: boolean
    /** Similar to readonly, but the input dims/greys out. */
    disabled?: boolean
    /** Optional tooltip to bind to input. */
    tooltip?: string
    /** Optional value. Should only be used for binding. */
    value?: any
    /** Optional rows for textarea. */
    rows?: number
    /** Optional min value for number input. */
    min?: number | undefined
    /** Optional max value for number input. */
    max?: number | undefined
    /** Optional options for select input. */
    options?: Option[] | undefined
    /** Optional input-box prefix attachment. */
    pre?: Snippet
    /** Optional input-box suffix attachment. */
    post?: Snippet
    /** Optional children to render inside the input. Useful for select options. */
    children?: Snippet
  }

  let {
    id,
    name = id,
    label = $_(`${id}.label`),
    placeholder = $bindable($_(`${id}.placeholder`)),
    description = $_(`${id}.description`),
    type = 'text',
    readonly = false,
    disabled = false,
    tooltip = $_(`${id}.tooltip`),
    value = $bindable(undefined),
    rows = 1,
    min = undefined,
    max = undefined,
    options = undefined,
    pre,
    children,
    post,
  }: Props = $props()

  type Option = { value: string | number | boolean; name: string; disabled?: boolean }

  let input = $state<SvelteComponent>()
  let showTooltip = $state(false)

  let currType = $derived(type)
  let passIcon = $derived(currType === 'password' ? faEyeSlash : faEye)
  $effect(() => {
    placeholder = placeholder == id + '.placeholder' ? '' : placeholder
  })

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
        <Button color="secondary" on:click={toggleTooltip} outline>
          {#if showTooltip}
            <Fa
              i={faArrowUpFromBracket}
              c1="gray"
              d1="gainsboro"
              c2="orange"
              scale="1.5x" />
          {:else}
            <Fa i={faQuestionCircle} c1="gray" d1="gainsboro" c2="orange" scale="1.5x" />
          {/if}
        </Button>
      {/if}
      {@render pre?.()}
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
          <!-- Create a boolean select-option list. -->
          <option value={false} selected={value === false}>
            {$_('words.select-option.Disabled')}
          </option>
          <option value={true} selected={value === true}>
            {$_('words.select-option.Enabled')}
          </option>
        {/if}
      </Input>

      <!-- Include a password visibility toggler. -->
      {#if type === 'password'}
        <Button type="button" outline on:click={togglePassword}>
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

    {#if showTooltip}
      <div transition:slide>
        <Card body class="mt-1" color="warning" outline>
          <p class="mb-0">{@html tooltip}</p>
        </Card>
      </div>
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
