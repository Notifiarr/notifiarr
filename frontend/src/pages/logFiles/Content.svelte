<script lang="ts">
  import { type BackendResponse, getUi } from '../../api/fetch'
  import type { LogFileInfo } from '../../api/notifiarrConfig'
  import Fa from '../../includes/Fa.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import {
    Button,
    Card,
    CardBody,
    Col,
    Input,
    InputGroup,
    ListGroup,
    ListGroupItem,
    Row,
  } from '@sveltestrap/sveltestrap'
  import {
    faSplotch as faColors,
    faSpinner,
    faQuestionCircle,
  } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import {
    faSplotch,
    faArrowDownShortWide,
    faArrowUpShortWide,
    faArrowUpFromBracket,
    faArrowProgress,
    faArrowRightArrowLeft,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { slide } from 'svelte/transition'
  import { delay, warning } from '../../includes/util'
  import { FileTail } from './tail.svelte'
  import { onMount } from 'svelte'
  import { get } from 'svelte/store'

  /** This module handles simple static content display,
   * and the tailing (active following) of files. */
  let { file, tail }: { file: LogFileInfo; tail?: boolean } = $props()

  // Form variables.
  // svelte-ignore non_reactive_update
  let lineCount = tail ? 50 : 500
  let offset = lineCount
  let desc = $state(true) // false == ascending (backward)
  let highlight = $state('')
  let colors = $state(true)
  let showTooltip = $state(false)
  let resp = $state<BackendResponse | FileTail>()
  let adding = $state(false)
  let loaded = $state(file.id)
  let wrap = $state(false)

  // Reload when a new file is selected.
  $effect(() => {
    if (loaded !== file.id) {
      loaded = file.id
      load(file.id)
    }
  })

  onMount(() => {
    // Initial load.
    load(file.id)
    return () => {
      // Disconnect websocket on unmount (nav away).
      if (resp instanceof FileTail) resp.destroy()
    }
  })

  function colorLine(line: string) {
    // User search highlight.
    if (highlight && line.includes(highlight)) return 'bg-success text-white'
    if (colors) {
      // This is a trigger/action.
      if (line.includes('requested]')) return 'bg-primary-subtle'
      // Services checks.
      if (line.includes('Critical')) return 'bg-warning-subtle'
      if (line.includes('DEBUG')) return 'bg-primary-subtle'
      // Catches any error. Might be too many.
      if (line.toLowerCase().includes('error')) return 'bg-danger-subtle'
      // Startup and info lines.
      if (line.includes('=>')) return 'bg-info-subtle'
      // Shutdown message(s).
      if (line.includes('!!>')) return 'bg-warning-subtle'
    }
    return ''
  }

  // Handles initial load, and reload button.
  const load = async (id: string | Event) => {
    if (id instanceof Event) {
      id.preventDefault()
      id = file.id
    }

    adding = true
    if (resp instanceof FileTail) {
      await resp.destroy()
      resp.body = get(_)('phrases.Loading')
      await delay(1000)
    }
    resp = undefined
    if (!tail) resp = await getUi(`getFile/logs/${id}/${lineCount}/0`, false)
    else resp = await new FileTail(file, lineCount)
    offset = lineCount
    adding = false
  }

  // Handles add more lines button.
  const add = async () => {
    adding = true
    const newR = await getUi(`getFile/logs/${file.id}/${lineCount}/${offset}`, false)
    if (newR.ok) {
      offset += lineCount
      resp ? (resp.body = newR.body + resp.body) : (resp = newR)
    } else {
      warning($_('LogFiles.Error', { values: { error: newR.body } }))
    }
    adding = false
  }
</script>

{#if !resp}
  <h3 class="text-success">
    <Fa i={faSpinner} spin scale={1.2} /> &nbsp; <T id="phrases.Loading" />
  </h3>
{:else if tail || resp.ok}
  {@const list = desc
    ? resp.body.trimEnd().split('\n')
    : resp.body.trimEnd().split('\n').reverse()}
  {@const lineNumberWidth = Math.floor(Math.log10(list.length)) + 1}
  <Row>
    <Col sm={12} md="auto" class="mb-2">
      <InputGroup style="width: auto !important;">
        <!-- Toggle Lines Order Button -->
        <Button
          outline
          onclick={() => (desc = !desc)}
          active={!desc}
          title={$_('LogFiles.ToggleLinesOrder')}>
          <Fa
            i={desc ? faArrowUpShortWide : faArrowDownShortWide}
            scale={1.5}
            c1="darkorange"
            c2="deeppink"
            d2="pink" />
        </Button>
        <!-- Toggle Colors Button -->
        <Button
          outline
          onclick={() => (colors = !colors)}
          active={!colors}
          title={$_('LogFiles.ToggleColors')}>
          <Fa
            spin={adding || (tail && resp.ok)}
            i={colors ? faColors : faSplotch}
            c1="purple"
            c2="violet"
            d1="gold"
            scale={1.5} />
        </Button>
        <!-- Line Count Input -->
        <Input
          title={$_('LogFiles.titles.LineCount')}
          style="width: 7rem !important;"
          type="number"
          min={10}
          max={10000}
          bind:value={lineCount} />
        <!-- Add / Following Button -->
        <Button
          outline
          onclick={add}
          title={$_('LogFiles.AddMoreLines')}
          disabled={file.used || adding || tail}>
          {#if !resp.ok}
            <b class="text-danger"><T id="phrases.ERROR" /></b>
          {:else if tail}
            <T id="LogFiles.titles.Tailing" />
          {:else}
            <T id="LogFiles.button.add" />
          {/if}
        </Button>
        <!-- Reload Button -->
        <Button outline onclick={load} title={$_('LogFiles.Reload')} disabled={adding}>
          <T id="LogFiles.button.reload" />
        </Button>
      </InputGroup>
    </Col>
    <Col class="mb-2">
      <InputGroup>
        <!-- Show More Button (Tooltip) -->
        <Button
          color="secondary"
          onclick={() => (showTooltip = !showTooltip)}
          outline
          style="width:44px;"
          title={$_('phrases.ShowMore')}>
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
        <!-- Highlight Input -->
        <Input bind:value={highlight} placeholder={$_('LogFiles.Highlight')} />
        <!-- Line Wrap Button -->
        <Button
          outline
          onclick={() => (wrap = !wrap)}
          aria-label={$_('LogFiles.ToggleLineWrap')}
          title={$_('LogFiles.ToggleLineWrap')}>
          <Fa
            i={wrap ? faArrowProgress : faArrowRightArrowLeft}
            c1="orange"
            d1="gold"
            c2="darkorange"
            d2="orange"
            scale={1.5} />
        </Button>
      </InputGroup>
    </Col>
    <Col xs={12}>
      {#if showTooltip}
        <div transition:slide>
          <Card body class="mt-1" color="warning" outline>
            <p class="mb-0"><T id="LogFiles.Tooltip" /></p>
          </Card>
        </div>
      {/if}
    </Col>
  </Row>

  <!-- File content is here. -->
  <div class="log-file-content" style="--line-number-width: {lineNumberWidth}ch;">
    <ListGroup flush numbered class="ps-0 overflow-auto overflow-y-hidden">
      {#each list as line}
        <ListGroupItem class="p-0 border-0 lh-1">
          <span
            class="d-inline-block {colorLine(line)}"
            style="margin-left: {lineNumberWidth + 1}ch;">
            <pre class="m-0 pre" class:wrap>{line}</pre>
          </span>
        </ListGroupItem>
      {/each}
    </ListGroup>
  </div>
{:else}
  <Card color="danger" outline>
    <CardBody><T id="LogFiles.Error" error={resp.body} /></CardBody>
  </Card>
{/if}

<style>
  /* All this to make the line numbers look good. */

  .log-file-content :global(.list-group) {
    counter-reset: liCounter;
  }

  .log-file-content :global(.list-group-item) {
    clear: both;
    position: relative;
  }

  .log-file-content :global(.list-group-item)::before {
    color: var(--bs-secondary-color);
    font-family: monospace;
    counter-increment: liCounter;
    content: counter(liCounter);
    display: inline-block;
    font-weight: 300;
    min-width: var(--line-number-width);
    text-align: right;
    position: absolute;
    left: 0;
    top: 0;
  }

  pre.pre {
    white-space: pre-wrap;
    word-break: break-all;
    word-wrap: break-word;
    scrollbar-width: none;
  }

  pre.wrap {
    white-space: pre !important;
    overflow: visible;
  }

  .log-file-content :global(.overflow-y-hidden) {
    overflow-y: hidden !important;
  }
</style>
