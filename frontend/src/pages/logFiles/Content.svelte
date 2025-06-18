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
</script>

{#await update}
  <h3 class="text-success">
    <Fa i={faSpinner} spin scale={1.2} /> &nbsp; <T id="phrases.Loading" />
  </h3>
{:then resp}
  {#if resp.ok}
    {@const list = resp.body.trimEnd().split('\n')}
    <div class="log-file-content">
      <ListGroup flush numbered class="ps-0 text-nowrap">
        {#each sort ? list : list.reverse() as line}
          <ListGroupItem class="p-0 border-0 lh-1">
            <span class="d-inline-block">
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
  .log-file-content :global(.list-group-item)::before {
    color: var(--bs-secondary-color);
    font-family: monospace;
  }

  pre.pre {
    white-space: pre-wrap;
    word-break: break-all;
    word-wrap: break-word;
  }
</style>
