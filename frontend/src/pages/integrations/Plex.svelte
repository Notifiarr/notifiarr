<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import type { PMSInfo, Sessions } from '../../api/notifiarrConfig'
  import { Card, Table, CardHeader } from '@sveltestrap/sveltestrap'
  import { age } from '../../includes/util'
  import Modal from './Modal.svelte'
  import { color, getLogo } from './data'
  import T, { _ } from '../../includes/Translate.svelte'
  import { getApi, type BackendResponse } from '../../api/fetch'
  import { faVideo } from '@fortawesome/sharp-duotone-light-svg-icons'
  import Nodal from '../../includes/Nodal.svelte'

  type Props = {
    status?: PMSInfo
    sessions?: Sessions
    plexAge?: Date
    sessionsAge?: Date
    showSessions?: boolean
    showOwner?: boolean
  }
  const {
    status,
    sessions,
    plexAge,
    sessionsAge,
    showSessions = true,
    showOwner = true,
  }: Props = $props()

  let statusModal: Modal | null = $state(null)
  let sessionsModal: Nodal | null = $state(null)
  let sessionsJson: Modal | null = $state(null)
  const app = 'plex'
</script>

<Card outline color={color(app)}>
  <CardHeader tag="div">
    <h5 class="m-0">
      <img
        src={getLogo(app)}
        alt="{app} Logo"
        class="float-start me-2"
        height="32"
        width="32" />
      {(sessions?.server ?? status?.friendlyName) || 'Plex'}
    </h5>
  </CardHeader>
  <Table class="rounded-bottom mb-0" size="sm">
    <tbody>
      {#if status?.friendlyName || sessions?.sessions?.length}
        <tr>
          <td class="text-nowrap"><T id="MediaApps.Plex.url.label" /></td>
          <td class="text-break">
            <a href={$profile.config.plex.url} target="_blank">
              {$profile.config.plex.url}</a>
          </td>
        </tr>
        {#if status}
          <tr>
            <td class="text-nowrap">
              <a href="#PlexStatus" onclick={statusModal?.toggle}>
                <T id="Integrations.titles.StatusAge" /></a>
              <Modal pageId="plexStatus" data={status} bind:this={statusModal} />
            </td>
            <td class="text-break">
              {age(profile.now - new Date(plexAge ?? 0).getTime())}</td>
          </tr>
          <tr>
            <td class="text-nowrap"><T id="Integrations.titles.Version" /></td>
            <td class="text-break">{status.version}</td>
          </tr>
          <tr>
            <td class="text-nowrap"><T id="Integrations.mediaTitles.PlexPass" /></td>
            <td class="text-break">{status.myPlexSubscription}</td>
          </tr>
          <tr>
            <td class="text-nowrap">
              <T id="Integrations.mediaTitles.PushNotifications" /></td>
            <td class="text-break">{status.pushNotifications}</td>
          </tr>
          {#if showOwner}
            <tr>
              <td class="text-nowrap"><T id="Integrations.mediaTitles.ServerOwner" /></td>
              <td class="text-break">{status.myPlexUsername}</td>
            </tr>
          {/if}
          <tr>
            <td class="text-nowrap"><T id="Integrations.mediaTitles.Country" /></td>
            <td class="text-break">{status.countryCode}</td>
          </tr>
          <tr>
            <td class="text-nowrap"><T id="system.OperatingSystem.Platform" /></td>
            <td class="text-break">{status.platform} {status.platformVersion}</td>
          </tr>
        {/if}
        {#if showSessions && sessions?.sessions?.length}
          <tr><td colspan="2"></td></tr>
          <tr>
            <td class="text-nowrap">
              <a href="#PlexSessionsJson" onclick={sessionsJson?.toggle}>
                <T id="Integrations.mediaTitles.SessionsAge" />
              </a>
              <Modal pageId="plexSessionsJson" data={sessions} bind:this={sessionsJson} />
            </td>
            <td class="text-break">
              {age(profile.now - new Date(sessionsAge ?? 0).getTime())}</td>
          </tr>
          <tr>
            <td class="text-nowrap"><T id="Integrations.mediaTitles.Sessions" /></td>
            <td class="text-break">
              <a href="#PlexSessions" onclick={sessionsModal?.open}>
                {sessions.sessions.length}</a>
            </td>
          </tr>
        {:else if showSessions}
          <tr>
            <td colspan="2"><T id="Integrations.phrases.NoCachedSessions" /></td>
          </tr>
        {/if}
      {:else}
        <tr>
          <td colspan="2">
            <b><T id="Integrations.phrases.NoPlexData" /></b>
          </td>
        </tr>
      {/if}
    </tbody>
  </Table>
</Card>

<Nodal
  title="Integrations.plexSessions"
  fa={{ i: faVideo }}
  get={() => getApi('plex/1/sessions')}
  size="xl"
  bind:this={sessionsModal}>
  {#snippet children(resp?: BackendResponse)}
    {@const sessions = (resp?.body?.message as Sessions) ?? {}}
    <Table striped bordered>
      <thead>
        <tr>
          <td colspan="2"><b><T id="Integrations.mediaTitles.User" /></b></td>
          <td><b><T id="Integrations.mediaTitles.Session" /></b></td>
          <td><b><T id="Integrations.mediaTitles.Complete" /></b></td>
          <td><b><T id="Integrations.mediaTitles.Encoding" /></b></td>
          <td><b><T id="Integrations.mediaTitles.Title" /></b></td>
        </tr>
      </thead>
      <tbody class="table-body">
        {#each sessions?.sessions ?? [] as session}
          <tr>
            <td title="uid: {session.User.id}" style="border-right: none;">
              <img
                referrerpolicy="no-referrer"
                src={session.User.thumb}
                height="25"
                width="25"
                alt={session.User.title} />
            </td>
            <td title="uid: {session.User.id}" style="border-left: none;">
              <h5>{session.User.title}</h5></td>
            <td>
              {session.type}
              <i class="text-{session.Player.state == 'paused' ? 'danger' : 'success'}">
                <T
                  id="Integrations.phrases.StateForTime"
                  state={session.Player.state}
                  timeDuration={age(
                    (session.Player.stateTime as unknown as number) * 1000,
                  )} />
              </i>
            </td>
            <td>
              {session.Player.title},
              {((session.viewOffset / session.duration) * 100).toFixed(1)}%
            </td>

            <td>
              {#each session.Media ?? [] as media, idx}
                {media.container}({media.videoCodec} @ {media.videoResolution} / {media.videoFrameRate},
                {media.audioCodec} * {media.audioChannels}){#if idx < session.Media!.length - 1};{/if}
              {/each}
            </td>
            <td>
              {#if session.grandparentTitle}{session.grandparentTitle};{/if}
              {session.title}
              {#if session.grandparentTitle}
                - S{session.parentIndex.toString().padStart(2, '0')}E{session.index
                  .toString()
                  .padStart(2, '0')}
              {/if}
              {#if session.year}({session.year}){/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </Table>
  {/snippet}
  {#snippet footer(resp?: BackendResponse)}
    <T id="Integrations.plexSessions.footer" />
  {/snippet}
</Nodal>

<style>
  .table-body :global(tr:last-of-type td) {
    border-bottom: 0 !important;
  }
</style>
