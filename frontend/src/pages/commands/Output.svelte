<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    CardBody,
    CardHeader,
    Popover,
  } from '@sveltestrap/sveltestrap'
  import Fa from '../../includes/Fa.svelte'
  import {
    faArrowDownFromBracket,
    faArrowRotateRight,
    faArrowUpFromBracket,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { age, delay, failure, since } from '../../includes/util'
  import T from '../../includes/Translate.svelte'
  import { theme } from '../../includes/theme.svelte'
  import type { Command, Stats } from '../../api/notifiarrConfig'
  import { getUi } from '../../api/fetch'
  import { slide } from 'svelte/transition'
  import { profile } from '../../api/profile.svelte'

  type Props = { index: number; active: boolean; form: Command }
  let { index, active, form }: Props = $props()

  let output = $state<Stats | null>()
  /** Whether the output is visible. */
  let showOutput = $state(false)
  /** The last time the output data was refreshed. */
  let lastRefreshed = $state(Date.now())
  /** How long ago the output data was refreshed. Dynamically updated. */
  let timeDuration = $state('')
  /** Toggle the output visibility. */
  const toggleOutput = (e?: Event) => (e?.preventDefault(), (showOutput = !showOutput))

  /** Retrieve the command output and update the last refreshed time. */
  export const getStats = async (e?: Event) => {
    e?.preventDefault()
    if (!form.hash) return
    const res = await getUi('ajax/cmdstats/' + form.hash)
    if (!res.ok) failure('Failure getting command output. ' + res.body)
    else output = res.body as Stats
    lastRefreshed = Date.now()
  }

  $effect(() => {
    // The delay prevents an error when saving the page.
    if (active) delay(50).then(() => getStats())
  })

  $effect(() => {
    // How long ago the output data was refreshed.
    timeDuration = age(profile.now - lastRefreshed)
  })

  /** A slide duration between 300ms and 2500ms based on the number of lines in the output. */
  let duration = $derived(
    Math.min(Math.max((output?.output.split('\n').length ?? 1) * 30, 300), 2500),
  )
</script>

<!-- These are easier to look at without so much indenting. -->
{#snippet buttons()}
  <Badge class="refresh"><T id="Commands.refreshed" {timeDuration} /></Badge>
  <Button id="outputRefresh{index}" color="success" size="sm" outline onclick={getStats}>
    <Fa i={faArrowRotateRight} c1="limegreen" c2="darkcyan" d2="cyan" scale="1.5" />
  </Button>
  <Button id="toggleCmd{index}" color="warning" size="sm" outline onclick={toggleOutput}>
    {#if showOutput}
      <Fa i={faArrowUpFromBracket} c1="orange" c2="brown" d2="wheat" scale="1.5" />
    {:else}
      <Fa i={faArrowDownFromBracket} c1="orange" c2="brown" d2="wheat" scale="1.5" />
    {/if}
  </Button>
  <!-- Refresh and output toggle buttons tooltips. -->
  <Popover target="toggleCmd{index}" trigger="hover" theme={$theme}>
    {#if showOutput}<T id="buttons.HideOutput" />
    {:else}<T id="buttons.ShowOutput" />
    {/if}
  </Popover>
  <Popover target="outputRefresh{index}" trigger="hover" theme={$theme}>
    <T id="buttons.RefreshOutput" />
  </Popover>
{/snippet}

<!-- The output loads when the page does, so this is always visible. -->
<Card class="mb-2">
  {#if output}
    <CardHeader>
      <h5 class="output-header mb-0">
        <T id="Commands.LastExecution" />
        <!-- Refresh and output toggle buttons. -->
        <div class="float-end">{@render buttons()}</div>
      </h5>
      {#if output.last == 'never'}
        <span class="text-muted"><T id="Commands.neverExecuted" /></span>
      {:else}
        <!-- Run stats. -->
        <T id="words.clock.ago" timeDuration={since(output.lastTime)} />
        &nbsp;<b class="text-secondary">|</b>&nbsp;
        <T id="Commands.Executions" count={output.runs} />
        &nbsp;<b class="text-secondary">|</b>&nbsp;
        <T id="Commands.Failures" count={output.fails} />
      {/if}
    </CardHeader>
    <!-- Command output content. -->
    <CardBody>
      <div class="text-primary font-monospace">{output.lastCmd}</div>
      {#if showOutput && output.output == ''}
        <div transition:slide={{ duration }} class="fst-italic fw-light">
          <T id="Commands.noOutput" />
        </div>
      {:else if showOutput}
        <pre class="pre" transition:slide={{ duration }}>{output.output}</pre>
      {/if}
    </CardBody>
  {/if}
</Card>

<style>
  .pre {
    white-space: pre-wrap;
    word-break: break-all;
    margin-bottom: 0;
  }

  .output-header :global(button) {
    margin-left: 0.5rem;
    width: 32px;
    height: 32px;
  }

  .output-header :global(.refresh) {
    margin-right: 3rem;
    margin-left: 0.5rem;
    font-size: 10px;
    font-weight: 400;
  }
</style>
