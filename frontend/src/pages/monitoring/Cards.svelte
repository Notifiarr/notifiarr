<script lang="ts">
  import {
    Table,
    CardHeader,
    Card,
    CardFooter,
    Button,
    Col,
    Popover,
    ButtonGroup,
  } from '@sveltestrap/sveltestrap'
  import T, { _, datetime } from '../../includes/Translate.svelte'
  import { age, since } from '../../includes/util'
  import { profile } from '../../api/profile.svelte'
  import { Monitor as chk } from './page.svelte'
  import { slide } from 'svelte/transition'
  import Fa from '../../includes/Fa.svelte'
  import {
    faArrowDownToBracket,
    faArrowUpFromBracket,
    faRedo,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { theme } from '../../includes/theme.svelte'

  type Props = { showOutput: Record<string, boolean> }
  const { showOutput = $bindable() }: Props = $props()
</script>

<div class="row cards-page">
  {#each chk.config?.results?.toSorted( (a, b) => a.name.localeCompare(b.name), ) ?? [] as result}
    {@const id = btoa(result.name + 'card').replace(/=/g, '')}
    {@const icon = showOutput[id] ? faArrowUpFromBracket : faArrowDownToBracket}

    <Col md="6" lg="6" xl="4" xxl="3">
      <Card color={chk.colors[result.state]} outline class="mt-3">
        <CardHeader>
          <h5 class="d-inline-block">{result.name}</h5>
          <Popover
            theme={$theme}
            target={id + 'state'}
            hideOnOutsideClick
            title={$_('monitoring.check.state.title')}>
            {#each Object.entries($profile.expvar.serviceChecks ?? {}) as [service, counts]}
              {#if service === result.name}
                {#each Object.entries(counts ?? {}).sort((a, b) => b[1] - a[1]) as [type, count]}
                  {type}: {count}<br />
                {/each}
              {/if}
            {/each}
          </Popover>
          <Button
            id={id + 'state'}
            color={chk.colors[result.state]}
            class="badge py-0 px-1">
            {chk.states[result.state]}</Button>
          <ButtonGroup class="float-end" size="sm" style="margin-right: -10px;">
            <Button
              color="primary"
              outline
              style="width: 2rem;"
              onclick={() => (showOutput[id] = !showOutput[id])}>
              <Fa i={icon} c1="blue" d1="slateblue" scale="1.4" />
            </Button>
            <Button
              color="primary"
              outline
              disabled={chk.checking[result.name]}
              onclick={e => chk.check(e, result.name)}>
              <Fa
                i={faRedo}
                spin={chk.checking[result.name]}
                c1="seagreen"
                c2="limegreen"
                scale="1.4" />
            </Button>
          </ButtonGroup>
          <div style="clear:both"></div>
        </CardHeader>

        <Table class="mb-0" size="sm" borderless striped>
          <tbody class="fit">
            <tr>
              <th>
                <Popover
                  theme={$theme}
                  hideOnOutsideClick
                  target={id + 'last'}
                  title={$_('monitoring.check.last.title', {
                    values: { timeDuration: since(result.time) },
                  })}>
                  {$_('monitoring.check.last.description')}
                  <h6>{datetime(result.time)}</h6>
                  {new Date(result.time).toISOString()}
                </Popover>
                <a href="#tooltip" id={id + 'last'} onclick={e => e.preventDefault()}>
                  {$_('monitoring.check.last.short')}
                </a>
              </th>
              <td>{since(result.time)}</td>
            </tr>

            <tr>
              <th>
                <Popover
                  theme={$theme}
                  hideOnOutsideClick
                  target={id + 'since'}
                  title={$_('monitoring.check.since.title', {
                    values: { timeDuration: since(result.since) },
                  })}>
                  {$_('monitoring.check.since.description')}
                  <h6>{datetime(result.since)}</h6>
                  {new Date(result.since).toISOString()}
                </Popover>
                <a href="#tooltip" id={id + 'since'} onclick={e => e.preventDefault()}>
                  {$_('monitoring.check.since.short')}
                </a>
              </th>
              <td>{since(result.since)}</td>
            </tr>

            <tr>
              <th>
                <Popover
                  theme={$theme}
                  hideOnOutsideClick
                  target={id + 'interval'}
                  title={$_('monitoring.check.interval.title')}>
                  {$_('monitoring.check.interval.tooltip')}
                </Popover>
                <a href="#tooltip" id={id + 'interval'} onclick={e => e.preventDefault()}>
                  {$_('monitoring.check.interval.short')}
                </a>
              </th>
              {#if result.interval > 0}
                <td>{age(result.interval * 1000)}</td>
              {:else}
                <td><T id="words.select-option.Disabled" /></td>
              {/if}
            </tr>
          </tbody>
        </Table>

        {#if showOutput[id]}
          <div transition:slide={{ duration: 320 }}>
            <CardFooter>
              <span class="text-muted">
                {$_('monitoring.check.typeExpect')}: {result.type}</span>
              <pre class="pre">{result.output}</pre>
            </CardFooter>
          </div>
        {:else if showOutput[id] === undefined}
          <!-- set this so the parent module can unset it with the toggle-all button -->
          {(showOutput[id] = false) || ''}
        {/if}
      </Card>
    </Col>
  {/each}
</div>

<style>
  tbody th a {
    text-decoration: none !important;
  }

  tbody.fit th {
    padding-left: 0.5rem;
  }

  /* Make the left-side headers fit the max-content length. */
  tbody.fit td {
    white-space: nowrap;
    width: 1%;
    padding-right: 1rem;
  }

  .pre {
    white-space: pre-wrap;
    word-break: break-all;
    margin-bottom: 0;
  }

  /* Small badge positioned to top. */
  .cards-page :global(.badge) {
    margin-left: 0.3rem;
    font-size: 10px;
    vertical-align: top;
  }
</style>
