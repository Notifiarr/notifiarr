<script lang="ts">
  import { Table, Popover } from '@sveltestrap/sveltestrap'
  import T, { _, datetime } from '../../includes/Translate.svelte'
  import { faQuestionCircle } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { faRedo } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from '../../includes/Fa.svelte'
  import { age, since } from '../../includes/util'
  import { profile } from '../../api/profile.svelte'
  import { Monitor as chk } from './page.svelte'
  import { theme } from '../../includes/theme.svelte'
</script>

<Table bordered striped responsive>
  <thead>
    <tr>
      <th>{$_('monitoring.check.name')}</th>
      <th class="fit">{$_('monitoring.check.state.short')}</th>
      <th class="fit">{$_('monitoring.check.typeExpect')}</th>
      <th class="fit">
        <Popover theme={$theme} hideOnOutsideClick target="tablelast">
          <div slot="title">{$_('monitoring.check.last.short')}</div>
          {$_('monitoring.check.last.tooltip')}
        </Popover>
        <span id="tablelast" class="help-icon">
          <Fa i={faQuestionCircle} c1="gray" d1="gainsboro" c2="orange" />
        </span>
        {$_('monitoring.check.last.short')}
      </th>
      <th class="fit">
        <Popover
          theme={$theme}
          hideOnOutsideClick
          target="tablesince"
          title={$_('monitoring.check.since.short')}>
          {$_('monitoring.check.since.tooltip')}
        </Popover>
        <span id="tablesince" class="help-icon">
          <Fa i={faQuestionCircle} c1="gray" d1="gainsboro" c2="orange" />
        </span>
        {$_('monitoring.check.since.short')}
      </th>
      <th class="fit">
        <Popover
          theme={$theme}
          hideOnOutsideClick
          target="tableinterval"
          title={$_('monitoring.check.interval.title')}>
          {$_('monitoring.check.interval.tooltip')}
        </Popover>
        <span id="tableinterval" class="help-icon">
          <Fa i={faQuestionCircle} c1="gray" d1="gainsboro" c2="orange" />
        </span>
        {$_('monitoring.check.interval.short')}
      </th>
      <th>{$_('monitoring.check.output')}</th>
    </tr>
  </thead>

  <tbody>
    {#each $profile.checkResults ?? [] as result}
      {@const id = btoa(result.name + 'table').replace(/=/g, '')}
      <tr>
        <td class="fw-bold">
          <a href="recheck/{result.name}" onclick={e => chk.check(e, result.name)}>
            <Fa i={faRedo} spin={chk.checking[result.name]} />
          </a>&nbsp;
          {result.name}
        </td>

        <td class="fit bg-{chk.colors[result.state]} text-white">
          {chk.states[result.state]}
          <Popover
            target={id + 'state'}
            theme={$theme}
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
          <span id={id + 'state'} class="help-icon">
            <Fa i={faQuestionCircle} c1="gainsboro" d1="gainsboro" c2="orange" />
          </span>
        </td>

        <td>{result.type}</td>

        <td class="fit">
          {since(result.time)}
          <Popover
            target={id + 'last'}
            theme={$theme}
            hideOnOutsideClick
            title={$_('monitoring.check.last.title', {
              values: { timeDuration: since(result.time) },
            })}>
            {$_('monitoring.check.last.description')}
            <h6>{datetime(result.time)}</h6>
            {new Date(result.time).toISOString()}
          </Popover>
          <span id={id + 'last'} class="help-icon">
            <Fa i={faQuestionCircle} c1="gray" d1="gainsboro" c2="orange" />
          </span>
        </td>

        <td class="fit">
          {since(result.since)}
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
          <span id={id + 'since'} class="help-icon">
            <Fa i={faQuestionCircle} c1="gray" d1="gainsboro" c2="orange" />
          </span>
        </td>

        {#if result.interval > 0}
          <td>{age(result.interval * 1000)}</td>
        {:else}
          <td><T id="words.select-option.Disabled" /></td>
        {/if}
        <td>{result.output?.toString() ?? ''}</td>
      </tr>
    {/each}
  </tbody>
</Table>

<style>
  .help-icon {
    cursor: pointer;
    margin-right: 0.2rem;
  }
  /* Make the left-side headers fit the max-content length. */
  thead th.fit,
  tbody td.fit {
    white-space: nowrap;
    width: 1%;
  }
</style>
