<script lang="ts">
  import {
    Card,
    Nav,
    NavItem,
    NavLink,
    Container,
    Row,
    Col,
    Button,
  } from '@sveltestrap/sveltestrap'
  import { profile } from './lib/login'
  import Configuration from './configuration/Index.svelte'
  import SiteTunnel from './siteTunnel/Index.svelte'
  import StarrApps from './starrApps/Index.svelte'
  import DownloadApps from './downloadApps/Index.svelte'
  import MediaApps from './mediaApps/Index.svelte'
  import SnapshotApps from './snapshotApps/Index.svelte'
  import FileWatcher from './fileWatcher/Index.svelte'
  import Endpoints from './endpoints/Index.svelte'
  import Commands from './commands/Index.svelte'
  import ServiceChecks from './serviceChecks/Index.svelte'
  import Triggers from './triggers/Index.svelte'
  import Integrations from './integrations/Index.svelte'
  import Monitoring from './monitoring/Index.svelte'
  import Metrics from './metrics/Index.svelte'
  import LogFiles from './logFiles/Index.svelte'
  import System from './system/Index.svelte'
  import Profile from './profile/Index.svelte'
  import Landing from './Landing.svelte'
  import { trimPrefix } from './lib/util'

  // Page structure for navigation with icons
  const settings = [
    { component: Configuration, id: 'configuration', name: 'Configuration', icon: '⚙️' },
    { component: SiteTunnel, id: 'siteTunnel', name: 'Site Tunnel', icon: '🔐' },
    { component: StarrApps, id: 'starrApps', name: 'Starr Apps', icon: '✨' },
    { component: DownloadApps, id: 'downloadApps', name: 'Downloaders', icon: '📥' },
    { component: MediaApps, id: 'mediaApps', name: 'Media Apps', icon: '🎬' },
    { component: SnapshotApps, id: 'snapshotApps', name: 'Snapshot', icon: '📸' },
    { component: FileWatcher, id: 'fileWatcher', name: 'File Watcher', icon: '👁️' },
    { component: Endpoints, id: 'endpoints', name: 'Endpoints', icon: '🔌' },
    { component: Commands, id: 'commands', name: 'Commands', icon: '🖥️' },
    { component: ServiceChecks, id: 'serviceChecks', name: 'Services', icon: '✓' },
  ]

  const insights = [
    { component: Triggers, id: 'triggers', name: 'Triggers', icon: '⚡' },
    { component: Integrations, id: 'integrations', name: 'Integrations', icon: '🔗' },
    { component: Monitoring, id: 'monitoring', name: 'Monitoring', icon: '📊' },
    { component: Metrics, id: 'metrics', name: 'Metrics', icon: '📈' },
    { component: LogFiles, id: 'logFiles', name: 'Log Files', icon: '📝' },
    { component: System, id: 'system', name: 'System', icon: '🖧' },
  ]

  const others = [
    { component: Profile, id: 'profile', name: 'Profile', icon: '👤' },
    { component: Landing, id: 'landing', name: 'Landing', icon: '🏠' },
  ]

  function navigateTo(event: Event, pageId: string): void {
    event.preventDefault()
    activePage = pageId
    window.history.replaceState({}, '', `${urlBase}${pageId}`)
    // Auto-collapse sidebar on mobile after navigation
    isOpen = innerWidth > 767
  }

  // Used to auto-navigate.
  $: urlBase = $profile?.urlBase || '/'
  $: parts = trimPrefix(window.location.pathname, urlBase).split('/')
  $: activePage = parts.length > 0 ? parts[0] : 'landing'

  // Used for sidebar collapse state.
  let innerWidth = 1200
  $: isOpen = innerWidth > 767

  // Set the component based on the active page. Dig it out of settings, others and insights.
  $: PageComponent =
    settings.concat(insights, others).find((page) => page.id === activePage)?.component || Landing
</script>

<svelte:window bind:innerWidth />

<div class="navigation">
  <Container fluid class="mt-3">
    <!-- Mobile Menu Toggle Button -->
    <div class="mobile-toggle-container d-md-none mb-3">
      <Button color="light" class="sidebar-toggle" on:click={() => (isOpen = !isOpen)}>
        {isOpen ? '✕ Hide' : '☰ Show'} Menu
      </Button>
    </div>

    <Row>
      <!-- Navigation Sidebar -->
      <div class={`sidebar-col col-md-3 col-lg-2 ${!isOpen ? 'd-none' : 'd-block'}`}>
        <Card body color="light" class="sidebar-card">
          <p class="navheader">Settings</p>
          <Nav vertical pills class="nav-custom">
            {#each settings as page}
              <NavItem>
                <NavLink
                  href={urlBase + page.id}
                  class="nav-link-custom"
                  active={activePage === page.id}
                  on:click={(e) => navigateTo(e, page.id)}
                >
                  <span class="nav-icon">{page.icon}</span>
                  <span class="nav-text">{page.name}</span>
                </NavLink>
              </NavItem>
            {/each}
          </Nav>

          <div class="section-divider"></div>

          <p class="navheader">Insights</p>
          <Nav vertical pills class="nav-custom">
            {#each insights as page}
              <NavItem>
                <NavLink
                  href={urlBase + page.id}
                  class="nav-link-custom"
                  active={activePage === page.id}
                  on:click={(e) => navigateTo(e, page.id)}
                >
                  <span class="nav-icon">{page.icon}</span>
                  <span class="nav-text">{page.name}</span>
                </NavLink>
              </NavItem>
            {/each}
          </Nav>

          <div class="user-profile">
            <div class="section-divider"></div>
            <Nav vertical pills class="nav-custom">
              <NavItem>
                <NavLink
                  href={urlBase + 'profile'}
                  class="nav-link-custom"
                  active={activePage === 'profile'}
                  on:click={(e) => navigateTo(e, 'profile')}
                >
                  <span class="profile-icon">{$profile?.username?.charAt(0).toUpperCase()}</span>
                  <span class="profile-name nav-text">{$profile?.username}</span>
                </NavLink>
              </NavItem>
            </Nav>
          </div>
        </Card>
      </div>

      <!-- Content Area -->
      <Col md={isOpen ? '9' : '12'} lg={10}>
        <svelte:component this={PageComponent} />
      </Col>
    </Row>
  </Container>
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

  .navigation :global(.nav-link-custom) {
    display: flex;
    align-items: center;
    padding: 8px 0px 8px 6px;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .navigation :global(.nav-link-custom.active) {
    background: linear-gradient(135deg, #1a73e8, #6c5ce7);
    box-shadow: 0 2px 5px rgba(108, 92, 231, 0.2);
  }

  .nav-icon {
    margin-right: 12px;
    font-size: 16px;
    display: inline-flex;
    width: 20px;
    justify-content: center;
  }

  .nav-text {
    font-size: 14px;
    font-weight: 500;
  }

  .user-profile {
    margin-top: auto;
    align-items: center;
    padding: 8px 0px;
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
    margin-right: 12px;
  }

  .profile-name {
    font-weight: 500;
    font-size: 14px;
  }

  /* Mobile styles */
  .mobile-toggle-container {
    position: sticky;
    top: 0;
    z-index: 1010;
    padding: 5px;
    background: rgba(255, 255, 255, 0.9);
    margin-bottom: 15px;
  }

  :global(.sidebar-toggle) {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    width: 100%;
    border-radius: 8px;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  }

  @media (max-width: 767.98px) {
    .navigation :global(.sidebar-col) {
      position: fixed;
      z-index: 1020;
      width: 85%;
      max-width: 230px;
      max-height: 100vh;
      overflow-y: auto;
      top: 0;
      left: 0;
      padding: 10px;
      background: rgb(203, 255, 237);
      transition: transform 0.3s ease;
      box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
    }
  }
</style>
