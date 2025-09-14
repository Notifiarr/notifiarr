<!-- This page is used to display the status of the Starr apps: Lidarr, Radarr, Readarr, Sonarr, and Prowlarr. -->
<script lang="ts">
  import type {
    RadarrSystemStatus,
    ReadarrSystemStatus,
    SonarrSystemStatus,
    StarrConfig,
    SystemStatus,
    ProwlarrSystemStatus,
    Queue,
    SonarrQueue,
    ReadarrQueue,
    RadarrQueue,
    State,
  } from '../../api/notifiarrConfig'
  import { age, formatBytes } from '../../includes/util'
  import { Card, CardHeader, Table } from '@sveltestrap/sveltestrap'
  import { color, getLogo, title } from './data'
  import Modal from './Modal.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import { profile } from '../../api/profile.svelte'

  type Props = {
    index: number
    config: StarrConfig
    app: 'lidarr' | 'radarr' | 'readarr' | 'sonarr' | 'prowlarr'
    status?:
      | SystemStatus
      | RadarrSystemStatus
      | ReadarrSystemStatus
      | SonarrSystemStatus
      | ProwlarrSystemStatus
    queue?: Queue | RadarrQueue | ReadarrQueue | SonarrQueue
    statusAge?: number | Date
    queueAge?: number | Date
    dashboard?: State[]
    dashboardAge?: number | Date
  }

  const {
    config,
    app,
    status,
    queue,
    index,
    statusAge,
    queueAge,
    dashboard,
    dashboardAge,
  }: Props = $props()

  let dashboardModal: Modal | null = $state(null)
  let statusModal: Modal | null = $state(null)
  let queueModal: Modal | null = $state(null)
</script>

<Card outline color={color(app)}>
  <CardHeader>
    <h5 class="m-0">
      <img
        src={getLogo(app)}
        alt="{app} Logo"
        class="float-start me-2"
        height="32"
        width="32" />
      {title(app)}
      {index + 1}
    </h5>
  </CardHeader>
  <Table class="rounded-bottom mb-0" size="sm">
    <tbody class="table-body">
      <tr>
        <td class="text-break">{config.name}</td>
        <td class="text-break"><a href={config.url}>{config.url}</a></td>
      </tr>
      {#if status}
        <tr>
          <td class="text-nowrap">
            <Modal pageId="{app}Status" data={status} bind:this={statusModal} />
            <a href="#{app}{index}Status" onclick={statusModal?.toggle}>
              <T id="Integrations.titles.StatusAge" /></a>
          </td>
          <td>{age(profile.now - new Date(statusAge ?? 0).getTime())}</td>
        </tr>
        <tr>
          <td class="text-nowrap"><T id="Integrations.titles.Version" /></td>
          <td class="text-break">{status?.version}</td>
        </tr>
        <tr>
          <td class="text-nowrap"><T id="Integrations.titles.Branch" /></td>
          <td class="text-break">{status?.branch}</td>
        </tr>
        <tr>
          <td class="text-nowrap"><T id="Integrations.titles.BuildTime" /></td>
          <td class="text-break">{status?.buildTime}</td>
        </tr>
      {/if}
      {#if queue && status}<tr><td colspan="2"></td></tr>{/if}
      {#if queue}
        <tr>
          <td class="text-nowrap">
            <Modal pageId="{app}Queue" data={queue} bind:this={queueModal} />
            <a href="#{app}{index}Queue" onclick={queueModal?.toggle}>
              <T id="Integrations.starrTitles.QueueAge" /></a>
          </td>
          <td class="text-break">
            {age(profile.now - new Date(queueAge ?? 0).getTime())}</td>
        </tr>
        <tr>
          <td class="text-nowrap"><T id="Integrations.starrTitles.QueueSize" /></td>
          <td class="text-break">{queue?.records?.length}</td>
        </tr>
      {/if}
      {#if dashboard && dashboard.length > 0}
        {#each dashboard as data}
          {#if data.instance == index + 1}
            <tr><td colspan="2"></td></tr>
            <tr>
              <td class="text-nowrap">
                <Modal pageId="{app}Dashboard" {data} bind:this={dashboardModal} />
                <a href="#{app}{index}Dashboard" onclick={dashboardModal?.toggle}>
                  <T id="Integrations.titles.DashboardAge" /></a>
              </td>
              <td class="text-break">
                {age(profile.now - new Date(dashboardAge ?? 0).getTime())}</td>
            </tr>
            <tr>
              <td class="text-nowrap"><T id="Integrations.starrTitles.Elapsed" /></td>
              <td class="text-break">{data.elapsed}</td>
            </tr>
            {#if app == 'sonarr'}
              <tr>
                <td class="text-nowrap"><T id="Integrations.starrTitles.Shows" /></td>
                <td class="text-break">{data.shows}</td>
              </tr>
              <tr>
                <td class="text-nowrap"><T id="Integrations.starrTitles.Episodes" /></td>
                <td class="text-break">{data.episodes}</td>
              </tr>
            {:else if app == 'radarr'}
              <tr>
                <td class="text-nowrap"><T id="Integrations.starrTitles.Movies" /></td>
                <td class="text-break">{data.movies}</td>
              </tr>
            {:else if app == 'readarr'}
              <tr>
                <td class="text-nowrap"><T id="Integrations.starrTitles.Books" /></td>
                <td class="text-break">{data.books}</td>
              </tr>
              <tr>
                <td class="text-nowrap"><T id="Integrations.starrTitles.Editions" /></td>
                <td class="text-break">{data.editions}</td>
              </tr>
            {:else if app == 'lidarr'}
              <tr>
                <td class="text-nowrap"><T id="Integrations.starrTitles.Tracks" /></td>
                <td class="text-break">{data.tracks}</td>
              </tr>
              <tr>
                <td class="text-nowrap"><T id="Integrations.starrTitles.Artists" /></td>
                <td class="text-break">{data.artists}</td>
              </tr>
            {/if}
            <tr>
              <td class="text-nowrap"><T id="Integrations.starrTitles.OnDisk" /></td>
              <td class="text-break">{data.onDisk}</td>
            </tr>
            <tr>
              <td class="text-nowrap"><T id="Integrations.starrTitles.Missing" /></td>
              <td class="text-break">{data.missing}</td>
            </tr>
            <tr>
              <td class="text-nowrap"><T id="Integrations.titles.Size" /></td>
              <td class="text-break">{formatBytes(data.size)}</td>
            </tr>
          {/if}
        {/each}
      {/if}
    </tbody>
  </Table>
</Card>

<style>
  .table-body :global(tr:last-of-type td) {
    border-bottom: 0 !important;
  }
</style>
