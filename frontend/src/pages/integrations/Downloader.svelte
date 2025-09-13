<!-- This page is used to display the status of the Downloader apps: Deluge, NZBGet, qBittorrent, rTorrent, and SABnzb. -->
<script lang="ts">
  import type {
    State,
    DelugeConfig,
    NZBGetConfig,
    QbitConfig,
    RtorrentConfig,
    SabNZBConfig,
  } from '../../api/notifiarrConfig'
  import T, { _ } from '../../includes/Translate.svelte'
  import { age, formatBytes } from '../../includes/util'
  import { Card, CardHeader, Table } from '@sveltestrap/sveltestrap'
  import { color, getLogo, title } from './data'
  import Modal from './Modal.svelte'
  import { profile } from '../../api/profile.svelte'

  type Props = {
    index: number
    config: DelugeConfig | NZBGetConfig | QbitConfig | RtorrentConfig | SabNZBConfig
    app: 'deluge' | 'nzbget' | 'qbit' | 'rtorrent' | 'sabnzbd' | 'transmission'
    dashboard?: State[]
    dashboardAge?: number | Date
  }

  const { config, app, index, dashboard, dashboardAge }: Props = $props()
  let dashboardModal: Modal | null = $state(null)
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

      {#if dashboard && dashboard.length > 0}
        {#each dashboard as data}
          {#if data.instance == index + 1}
            <tr>
              <td class="text-nowrap">
                <Modal pageId="{app}Dashboard" {data} bind:this={dashboardModal} />
                <a href="#{app}{index}Dashboard" onclick={dashboardModal?.toggle}>
                  <T id="Integrations.titles.DashboardAge" />
                </a>
              </td>
              <td class="text-break">
                {age(profile.now - new Date(dashboardAge ?? 0).getTime())}</td>
            </tr>
            <tr>
              <td class="text-nowrap"><T id="Integrations.starrTitles.Elapsed" /></td>
              <td class="text-break">{data.elapsed}</td>
            </tr>
            <tr>
              <td class="text-nowrap">
                <T id="Integrations.downloadTitles.Downloads" /></td>
              <td class="text-break">{data.downloads}</td>
            </tr>
            <tr>
              <td class="text-nowrap"
                ><T id="Integrations.downloadTitles.Incomplete" /></td>
              <td class="text-break">{data.incomplete}</td>
            </tr>
            <tr>
              <td class="text-nowrap">
                <T id="Integrations.downloadTitles.Downloading" /></td>
              <td class="text-break">{data.downloading}</td>
            </tr>
            <tr>
              <td class="text-nowrap"><T id="Integrations.downloadTitles.Paused" /></td>
              <td class="text-break">{data.paused}</td>
            </tr>
            <tr>
              <td class="text-nowrap">
                <T id="Integrations.downloadTitles.TotalSize" /></td>
              <td class="text-break">{formatBytes(data.size)}</td>
            </tr>
            <tr>
              <td class="text-nowrap"><T id="Integrations.downloadTitles.Errors" /></td>
              <td class="text-break">{data.errors}</td>
            </tr>
            <tr>
              <td class="text-nowrap">
                <T id="Integrations.downloadTitles.Downloaded" /></td>
              <td class="text-break">{formatBytes(data.downloaded ?? 0)}</td>
            </tr>
            {#if app == 'deluge' || app == 'qbit' || app == 'rtorrent'}
              <tr>
                <td class="text-nowrap">
                  <T id="Integrations.downloadTitles.Uploading" /></td>
                <td class="text-break">{data.uploading}</td>
              </tr>
              <tr>
                <td class="text-nowrap">
                  <T id="Integrations.downloadTitles.Seeding" /></td>
                <td class="text-break">{data.seeding}</td>
              </tr>
              <tr>
                <td class="text-nowrap">
                  <T id="Integrations.downloadTitles.Uploaded" /></td>
                <td class="text-break">{formatBytes(data.uploaded ?? 0)}</td>
              </tr>
            {:else if app == 'sabnzbd'}
              <tr>
                <td class="text-nowrap"><T id="Integrations.downloadTitles.Month" /></td>
                <td class="text-break">{formatBytes(data.month ?? 0)}</td>
              </tr>
              <tr>
                <td class="text-nowrap"><T id="Integrations.downloadTitles.Week" /></td>
                <td class="text-break">{formatBytes(data.week ?? 0)}</td>
              </tr>
              <tr>
                <td class="text-nowrap border-0">
                  <T id="Integrations.downloadTitles.Day" /></td>
                <td class="text-break border-0">{formatBytes(data.day ?? 0)}</td>
              </tr>
            {/if}
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
