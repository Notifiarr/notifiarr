<script lang="ts" module>
  import { urlbase } from './api/fetch'
  import { get, writable } from 'svelte/store'

  // Page structure for navigation with icons
  // 'id' (from page) is used for navigation AND translations.

  // Settings header in navigation menu.
  const settings = [
    { component: Configuration, ...ConfigP },
    { component: SiteTunnel, ...SiteTunnelP },
    { component: StarrApps, ...StarrAppsP },
    { component: DownloadApps, ...DownloadAppsP },
    { component: MediaApps, ...MediaAppsP },
    { component: SnapshotApps, ...SnapshotAppsP },
    { component: FileWatcher, ...FileWatcherPage },
    { component: Endpoints, ...EndpointsPage },
    { component: Commands, ...CommandsPage },
    { component: ServiceChecks, ...ServicesP },
  ]
  // Insights header in navigation menu.
  const insights = [
    { component: Triggers, ...TriggersP },
    { component: Integrations, ...IntegrationsP },
    { component: Monitoring, ...MonitoringP },
    { component: Metrics, ...MetricsP },
    { component: LogFiles, ...LogFilesP },
    { component: System, ...SystemP },
  ]
  // Others do not show up in the navigation menu.
  const others = [
    { component: Profile, ...ProfilePage },
    { component: Landing, ...LandingPage },
  ]

  // Used for sidebar collapse state.
  let windowWidth = $state(1000)
  let isOpen = $derived(windowWidth >= 992)
  // Used to navigate.
  const activePage = writable('')
  let ActivePage = $state(Landing as Component)

  /**
   * Used to navigate to a page.
   * @param event - from an onclick handler, optional.
   * @param pid - the id of the page to navigate to, ie profile, configuration, etc.
   */
  export function goto(event: Event | null, pid: string, subPages: string[] = []): void {
    event?.preventDefault()
    navTo(new PopStateEvent('popstate', { state: { uri: pid } }))
    const query = new URLSearchParams(window.location.search)
    const params = query.toString()
    const uri = `${get(urlbase)}${[pid, ...subPages].join('/').toLowerCase()}${params ? `?${params}` : ''}`
    window.history.pushState({ uri: get(activePage) }, '', uri)
  }

  // navTo is split from goto(), so we can call it from popstate.
  // Call this on load and when the back button is clicked.
  function navTo(e: PopStateEvent) {
    e.preventDefault()
    const newPage = e.state?.uri ?? ''
    const page = settings.concat(insights, others).find(p => iequals(p.id, newPage))
    activePage.set(page ? newPage : '')
    ActivePage = page?.component || Landing
  }
</script>

<script lang="ts">
  import {
    Card,
    Nav,
    NavItem,
    NavLink,
    Row,
    Col,
    Button,
    Dropdown,
    DropdownToggle,
    DropdownMenu,
    DropdownItem,
    Input,
    Theme,
  } from '@sveltestrap/sveltestrap'
  import { profile } from './api/profile.svelte'
  import { _ } from './includes/Translate.svelte'
  import Configuration, { page as ConfigP } from './pages/configuration/Index.svelte'
  import SiteTunnel, { page as SiteTunnelP } from './pages/siteTunnel/Index.svelte'
  import StarrApps, { page as StarrAppsP } from './pages/starrApps/Index.svelte'
  import DownloadApps, { page as DownloadAppsP } from './pages/downloadApps/Index.svelte'
  import MediaApps, { page as MediaAppsP } from './pages/mediaApps/Index.svelte'
  import SnapshotApps, { page as SnapshotAppsP } from './pages/snapshotApps/Index.svelte'
  import FileWatcher, { page as FileWatcherPage } from './pages/fileWatcher/Index.svelte'
  import Endpoints, { page as EndpointsPage } from './pages/endpoints/Index.svelte'
  import Commands, { page as CommandsPage } from './pages/commands/Index.svelte'
  import ServiceChecks, { page as ServicesP } from './pages/serviceChecks/Index.svelte'
  import Triggers, { page as TriggersP } from './pages/triggers/Index.svelte'
  import Integrations, { page as IntegrationsP } from './pages/integrations/Index.svelte'
  import Monitoring, { page as MonitoringP } from './pages/monitoring/Index.svelte'
  import Metrics, { page as MetricsP } from './pages/metrics/Index.svelte'
  import LogFiles, { page as LogFilesP } from './pages/logFiles/Index.svelte'
  import System, { page as SystemP } from './pages/system/Index.svelte'
  import Profile, { page as ProfilePage } from './pages/profile/Index.svelte'
  import Landing, { page as LandingPage } from './Landing.svelte'
  import { iequals, ltrim } from './includes/util'
  import { theme as thm } from './includes/theme.svelte'
  import { currentLocale, setLocale } from './includes/locale/index.svelte'
  import { Flags } from './includes/locale/index.svelte'
  import { faStarship, faSun } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import Fa from './includes/Fa.svelte'
  import { slide } from 'svelte/transition'
  import { onMount, type Component } from 'svelte'

  let theme = $derived($thm)
  let newLang = $derived(currentLocale())

  onMount(() => {
    // Navigate to the initial page based on the URL when the content mounts.
    const parts = ltrim(window.location.pathname, get(urlbase)).split('/')
    const state = { uri: parts.length > 0 ? parts[0] : '' }
    navTo(new PopStateEvent('popstate', { state }))
  })
</script>

<svelte:window bind:innerWidth={windowWidth} on:popstate={navTo} />

<div class="navigation">
  <!-- Mobile Menu Toggle Button -->
  <Card color="warning" {theme} class="menu-toggle-wrapper d-lg-none mb-2 p-0" outline>
    <Button size="sm" class="menu-toggle-button my-0" onclick={() => (isOpen = !isOpen)}>
      {isOpen ? `✕ ${$_('phrases.HideMenu')}` : `☰ ${$_('phrases.ShowMenu')}`}
    </Button>
  </Card>

  <Row>
    <!-- Navigation Sidebar -->
    {#if isOpen}
      <div class="sidebar-col col mb-2" transition:slide={{ duration: 475, axis: 'x' }}>
        <Card body class="sidebar-card pb-2" {theme}>
          <!-- Settings -->
          <p class="navheader">{$_('navigation.titles.Settings')}</p>
          <Nav vertical pills class="nav-custom" {theme}>
            {#each settings as page}
              <NavItem>
                <NavLink
                  href={$urlbase + page.id}
                  class="nav-link-custom"
                  active={iequals($activePage, page.id)}
                  disabled={iequals($activePage, page.id)}
                  onclick={e => goto(e, page.id)}>
                  <span class="nav-icon">
                    <Fa {...page} scale="1.7x" />
                  </span>
                  <span class="nav-text">{$_('navigation.titles.' + page.id)}</span>
                </NavLink>
              </NavItem>
            {/each}
          </Nav>

          <div class="section-divider"></div>

          <!-- Insights -->
          <p class="navheader">{$_('navigation.titles.Insights')}</p>
          <Nav vertical pills class="nav-custom" {theme}>
            {#each insights as page}
              <NavItem>
                <NavLink
                  href={$urlbase + page.id}
                  class="nav-link-custom"
                  active={iequals($activePage, page.id)}
                  disabled={iequals($activePage, page.id)}
                  onclick={e => goto(e, page.id)}>
                  <span class="nav-icon">
                    <Fa {...page} scale="1.7x" />
                  </span>
                  <span class="nav-text">{$_('navigation.titles.' + page.id)}</span>
                </NavLink>
              </NavItem>
            {/each}
          </Nav>

          <!-- Profile Dropdown -->
          <div class="mt-auto pt-2">
            <div class="section-divider"></div>
            <Nav vertical pills class="nav-custom" {theme}>
              <Dropdown nav direction="up" class="ms-0">
                <DropdownToggle
                  nav
                  class="dropdown-custom {$activePage.toLowerCase() === 'trustprofile'
                    ? 'active'
                    : ''}">
                  <span class="text-uppercase profile-icon">
                    {$profile.username.charAt(0).toUpperCase()}
                  </span>
                  <span>{$profile.username}</span>
                </DropdownToggle>
                <DropdownMenu>
                  <DropdownItem
                    class="nav-link-custom"
                    active={iequals($activePage, 'TrustProfile')}
                    disabled={iequals($activePage, 'TrustProfile')}
                    onclick={e => goto(e, 'TrustProfile')}>
                    <span class="nav-icon"><Fa {...ProfilePage} scale="1.7x" /></span>
                    <span class="nav-text">{$_('navigation.titles.TrustProfile')}</span>
                  </DropdownItem>
                  <Input
                    class="my-1 lang-select"
                    type="select"
                    bind:value={newLang}
                    onchange={() => setLocale(newLang)}>
                    {#each Object.entries($profile.languages?.[currentLocale()] || {}) as [code, lang]}
                      <option value={code} selected={code === currentLocale()}>
                        {Flags[code]}&nbsp;&nbsp; {lang.name}
                      </option>
                    {/each}
                  </Input>
                  <DropdownItem class="nav-link-custom" onclick={thm.toggle}>
                    <Fa
                      i={faStarship}
                      d={faSun}
                      c1="green"
                      c2="lightblue"
                      d1="orange"
                      d2="fuchsia"
                      scale="1.3x"
                      class="me-3" />
                    {theme.includes('dark')
                      ? $_('config.titles.Light')
                      : $_('config.titles.Dark')}
                    <Input
                      disabled
                      type="switch"
                      checked={theme.includes('dark')}
                      style="position: absolute; right: 5px;" />
                  </DropdownItem>
                </DropdownMenu>
              </Dropdown>
            </Nav>
          </div>
        </Card>
      </div>
    {/if}

    <!-- Content Area -->
    <Col>
      <Theme {theme}>
        <Card class="mb-2" outline color="notifiarr">
          <ActivePage />
        </Card>
      </Theme>
    </Col>
  </Row>
</div>

<style>
  .navigation :global(.sidebar-card) {
    position: sticky;
    top: 20px;
    border-radius: 12px;
    padding: 10px 5px 10px 5px;
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.05);
    min-height: calc(100vh - 150px);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
  }

  .section-divider {
    height: 1px;
    background-color: rgba(0, 0, 0, 0.1);
    margin: 12px 0;
    width: 100%;
  }

  .navheader {
    font-weight: 600;
    font-size: 12px;
    text-transform: uppercase;
    color: #666;
    margin-top: 5px;
    margin-bottom: 8px;
    padding-left: 8px;
    letter-spacing: 0.5px;
  }

  .navigation :global(.nav-custom) {
    gap: 4px;
  }

  .navigation :global(.dropdown-custom.active) {
    background: linear-gradient(135deg, #1a73e8, #6c5ce7);
    box-shadow: 0 2px 5px rgba(108, 92, 231, 0.2);
    transition: all 0.2s ease;
    border-radius: 8px;
    color: #f1f3f5;
  }

  .navigation :global(.nav-link-custom) {
    display: flex;
    align-items: center;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .navigation :global(.nav-link-custom.active) {
    background: linear-gradient(135deg, #1a73e8, #6c5ce7);
    box-shadow: 0 2px 5px rgba(108, 92, 231, 0.2);
  }

  .nav-icon {
    margin-right: 12px;
    display: inline-flex;
    width: 20px;
  }

  .profile-icon {
    width: 24px;
    height: 24px;
    background: #f1f3f5;
    color: #495057;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    margin-right: 6px;
  }

  /* Mobile styles */
  .navigation :global(.menu-toggle-wrapper) {
    position: sticky;
    top: 0;
    z-index: 1010;
    padding: 5px;
    margin-bottom: 15px;
    text-align: center;
    border-radius: 3px;
  }

  .navigation :global(.menu-toggle-button) {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    width: 100%;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  }

  .navigation :global(.sidebar-col) {
    min-width: 230px;
    max-width: 230px;
  }

  @media (max-width: 991.98px) {
    .navigation :global(.sidebar-col) {
      position: fixed;
      z-index: 1020;
      width: 85%;
      max-height: 100vh;
      overflow-y: auto;
      top: 0;
      left: 0;
      padding: 1px;
      border-radius: 12px;
      transition: transform 0.2s cubic-bezier(0.25, 0.1, 0.25, 1);
      box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
      background: rgba(118, 122, 126, 0.9);
    }
  }

  .navigation :global(.lang-select) {
    padding: 1px 0px 1px 14px;
    border: 0px;
    font-size: 16px;
  }
</style>
