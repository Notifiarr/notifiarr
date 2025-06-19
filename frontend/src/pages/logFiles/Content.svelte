<script lang="ts">
  import { faSpinner } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { getUi } from '../../api/fetch'
  import type { LogFileInfo } from '../../api/notifiarrConfig'
  import Fa from '../../includes/Fa.svelte'
  import { Card, CardBody, ListGroup, ListGroupItem } from '@sveltestrap/sveltestrap'
  import T from '../../includes/Translate.svelte'

  let { file }: { file: LogFileInfo } = $props()
  let lineCount = $state(500)
  let offset = $state(0)
  let sort = $state(true) // true == desc, false == asc

  const update = $derived.by(() =>
    getUi(`getFile/logs/${file.id}/${lineCount}/${offset}`, false),
  )

  function colorLine(line: string) {
    // This is a trigger/action.
    if (line.includes('requested]')) return 'primary-subtle'
    // Services checks.
    if (line.includes('Critical')) return 'warning-subtle'
    if (line.includes('DEBUG')) return 'primary-subtle'
    // Catches any error. Might be too many.
    if (line.toLowerCase().includes('error')) return 'danger-subtle'
    // Startup and info lines.
    if (line.includes('=>')) return 'info-subtle'
    // Shutdown message(s).
    if (line.includes('!!>')) return 'warning-subtle'
    return ''
  }
</script>

{#await update}
  <h3 class="text-success">
    <Fa i={faSpinner} spin scale={1.2} /> &nbsp; <T id="phrases.Loading" />
  </h3>
{:then resp}
  {#if resp.ok}
    {@const list = resp.body.trimEnd().split('\n')}
    {@const lineNumberWidth = Math.floor(Math.log10(list.length)) + 1 + 'ch'}
    <div class="log-file-content" style="--line-number-width: {lineNumberWidth}">
      <ListGroup flush numbered class="ps-0 text-nowrap ms-0">
        {#each sort ? list : list.reverse() as line}
          <ListGroupItem class="p-0 border-0 lh-1 ms-0">
            <span class="d-inline-block me-0 bg-{colorLine(line)}">
              <pre class="mb-0 me-4 pre">{line}</pre>
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
{/await}

<style>
  /* All this to make the line numbers look good. */

  .log-file-content :global(.list-group) {
    counter-reset: liCounter;
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
    margin-right: 0.4rem;
  }

  pre.pre {
    white-space: pre-wrap;
    word-break: break-all;
    word-wrap: break-word;
  }
</style>
