<script lang="ts">
  import { Table } from '@sveltestrap/sveltestrap'
  import T, { _, datetime } from '../../includes/Translate.svelte'
  import { profile } from '../../api/profile.svelte'
  import { since } from '../../includes/util'
</script>

<!-- Tunnel Stats Table at bottom of page.-->
<h4><T id="SiteTunnel.phrases.TunnelStats" /></h4>
<Table size="sm" striped>
  {#if $profile.tunnelPoolStats}
    <tbody class="fit">
      {#each Object.entries($profile.tunnelPoolStats) as [socket, stats]}
        <tr>
          <th>
            {#if stats?.Active}
              <T id="SiteTunnel.phrases.SocketURLActive" />
            {:else}
              <T id="SiteTunnel.phrases.SocketURLInactive" />
            {/if}
          </th>
          <td><b>{socket}</b></td>
        </tr>
        <tr>
          <th><T id="SiteTunnel.phrases.Disconnects" /></th>
          <td>{stats?.Disconnects}</td>
        </tr>
        <tr>
          <th><T id="SiteTunnel.phrases.ConnectionPoolSize" /></th>
          <td>{stats?.Total}</td>
        </tr>
        <tr>
          <th><T id="SiteTunnel.phrases.Connecting" /></th><td>{stats?.Connecting}</td>
        </tr>
        <tr> <th><T id="SiteTunnel.phrases.Idle" /></th><td>{stats?.Idle}</td> </tr>
        <tr>
          <th><T id="SiteTunnel.phrases.Running" /></th><td>{stats?.Running}</td>
        </tr>
        <tr>
          <th><T id="SiteTunnel.phrases.LastConnection" /></th>
          <td>
            {datetime(stats!.LastConn)},
            <T id="words.clock.ago" timeDuration={since(stats!.LastConn)} />
          </td>
        </tr>
        <tr>
          <th><T id="SiteTunnel.phrases.LastActiveCheck" /></th>
          <td>
            {datetime(stats!.LastTry)},
            <T id="words.clock.ago" timeDuration={since(stats!.LastTry)} />
          </td>
        </tr>
      {/each}
    </tbody>
  {/if}
</Table>
<span class="text-muted"><T id="SiteTunnel.phrases.TunnelStatsStale" /></span>

<style>
  /* Make the left-side headers fit the max-content length. */
  tbody.fit th {
    white-space: nowrap;
    width: 1%;
    padding-right: 1rem;
  }
</style>
