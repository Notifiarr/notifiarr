<script lang="ts" module>
  import { faLinkHorizontal } from '@fortawesome/sharp-duotone-solid-svg-icons'
  export const page = {
    id: 'Integrations',
    i: faLinkHorizontal,
    c1: 'purple',
    c2: 'gray',
    d1: 'magenta',
    d2: 'white',
  }
</script>

<script lang="ts">
  import { Alert, CardBody, Col, Row, Spinner } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import Header from '../../includes/Header.svelte'
  import { getUi } from '../../api/fetch'
  import { onMount } from 'svelte'
  import Plex from './Plex.svelte'
  import { type Config, type Integrations } from '../../api/notifiarrConfig'
  import Starr from './Starr.svelte'
  import { profile } from '../../api/profile.svelte'
  import Downloader from './Downloader.svelte'
  import Tautulli from './Tautulli.svelte'
  import Snapshot from './Snapshot/Index.svelte'

  let response: Promise<Integrations> | null = $state(null)

  const getIntegrations = async (): Promise<Integrations> => {
    const resp = await getUi('integrations')
    if (!resp.ok) throw new Error(resp.body)
    return resp.body as Integrations
  }

  onMount(() => {
    response = getIntegrations()
  })
</script>

{#snippet starr(
  resp: Integrations,
  config: Config,
  app: 'sonarr' | 'radarr' | 'readarr' | 'lidarr' | 'prowlarr',
)}
  {#each config[app] ?? [] as status, index}
    {#if config[app]}
      {@const queue = app == 'prowlarr' ? undefined : resp?.[app]?.queue?.[index]}
      {@const queueAge = app == 'prowlarr' ? undefined : resp?.[app]?.queueAge?.[index]}
      <!-- make a slider in the ui control these values -->
      <Col class="mb-2" sm={12} md={6} xxl={4}>
        <Starr
          {app}
          {index}
          {queue}
          {queueAge}
          config={config[app][index]}
          status={resp?.[app]?.status?.[index]}
          statusAge={resp?.[app]?.statusAge?.[index]}
          dashboard={app == 'prowlarr' ? undefined : resp?.dashboard?.[app]}
          dashboardAge={resp?.dashboardAge} />
      </Col>
    {/if}
  {/each}
{/snippet}

{#snippet downloader(
  resp: Integrations,
  config: Config,
  app: 'deluge' | 'nzbget' | 'qbit' | 'rtorrent' | 'sabnzbd' | 'transmission',
)}
  {#each config[app] ?? [] as status, index}
    {#if config[app]}
      <Col class="mb-2" sm={12} md={6} xxl={4}>
        <Downloader
          {app}
          {index}
          config={config[app][index]}
          dashboard={resp?.dashboard?.[app]}
          dashboardAge={resp?.dashboardAge} />
      </Col>
    {/if}
  {/each}
{/snippet}

{#snippet content(resp: Integrations, config: Config)}
  <Row><Col><h4 class="mt-0"><T id="navigation.titles.MediaApps" /></h4></Col></Row>
  <Row>
    <Col class="mb-2" sm={12} md={config.tautulli ? 6 : 12}>
      <Plex
        status={resp.plex}
        sessions={resp.sessions}
        plexAge={resp.plexAge}
        sessionsAge={resp.sessionsAge} />
    </Col>

    {#if config.tautulli}
      <Col class="mb-2" sm={12} md={6}>
        <Tautulli
          config={config.tautulli}
          status={resp.tautulli}
          users={resp.tautulliUsers?.response.data ?? []}
          statusAge={resp.tautulliAge}
          usersAge={resp.tautulliUsersAge} />
      </Col>
    {/if}
  </Row>

  <Row><Col><h4><T id="navigation.titles.StarrApps" /></h4></Col></Row>
  <Row>
    {#each ['sonarr', 'radarr', 'readarr', 'lidarr', 'prowlarr'] as app}
      {@render starr(
        resp,
        config,
        app as 'sonarr' | 'radarr' | 'readarr' | 'lidarr' | 'prowlarr',
      )}
    {/each}
  </Row>

  <Row><Col><h4><T id="navigation.titles.Downloaders" /></h4></Col></Row>
  <Row>
    {#each ['deluge', 'nzbget', 'qbit', 'rtorrent', 'sabnzbd', 'transmission'] as app}
      {@render downloader(
        resp,
        config,
        app as 'deluge' | 'nzbget' | 'qbit' | 'rtorrent' | 'sabnzbd' | 'transmission',
      )}
    {/each}
  </Row>

  <!-- Snapshot data -->
  {#if resp.snapshot}
    <Row><Col><h4><T id="Integrations.Snapshot.Latest" /></h4></Col></Row>
    <Row>
      <Snapshot snapshot={resp.snapshot} snapshotAge={resp.snapshotAge} />
    </Row>
  {/if}
{/snippet}

<Header {page} />

<CardBody>
  {#if response}
    {#await response}
      <Spinner /> {$_('phrases.Loading')}
    {:then resp}
      {@render content(resp, $profile.config)}
    {:catch err}
      <Alert color="danger">{err.message}</Alert>
    {/await}
  {/if}
</CardBody>
