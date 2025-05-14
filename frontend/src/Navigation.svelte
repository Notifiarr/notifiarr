<script lang="ts">
  import {
    Card,
    Nav,
    NavItem,
    NavLink,
    Row,
    Col,
    Button,
    Fade,
    Dropdown,
    DropdownToggle,
    DropdownMenu,
    DropdownItem,
    Input,
    Icon,
  } from '@sveltestrap/sveltestrap'
  import { profile } from './api/profile.svelte'
  import { _ } from './lib/Translate.svelte'
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
  import { ltrim } from './lib/util'
  import { darkMode, toggleDarkMode } from './lib/darkmode.svelte'
  import { currentLocale, setLocale } from './lib/locale/index.svelte'
  import { Flags } from './lib/locale/index.svelte'
  import { urlbase } from './api/fetch'

  $: theme = $darkMode ? 'dark' : 'light'
  // Page structure for navigation with icons
  // id used for navigation AND translations.
  const settings = [
    { component: Configuration, id: 'Configuration', icon: '⚙️' },
    { component: SiteTunnel, id: 'SiteTunnel', icon: '🔐' },
    { component: StarrApps, id: 'StarrApps', icon: '✨' },
    { component: DownloadApps, id: 'Downloaders', icon: '📥' },
    { component: MediaApps, id: 'MediaApps', icon: '🎬' },
    { component: SnapshotApps, id: 'SnapshotApps', icon: '📸' },
    { component: FileWatcher, id: 'FileWatcher', icon: '👁️' },
    { component: Endpoints, id: 'Endpoints', icon: '🔌' },
    { component: Commands, id: 'Commands', icon: '🖥️' },
    { component: ServiceChecks, id: 'Services', icon: '✓' },
  ]

  const insights = [
    { component: Triggers, id: 'Triggers', icon: '⚡' },
    { component: Integrations, id: 'Integrations', icon: '🔗' },
    { component: Monitoring, id: 'Monitoring', icon: '📊' },
    { component: Metrics, id: 'Metrics', icon: '📈' },
    { component: LogFiles, id: 'LogFiles', icon: '📝' },
    { component: System, id: 'System', icon: '🖧' },
  ]

  const others = [
    { component: Profile, id: 'TrustProfile', icon: '👤' },
    { component: Landing, id: '', icon: '🏠' },
  ]

  /**
   * Used to navigate to a page.
   * @param event - from an onclick handler
   * @param pageId - the id of the page to navigate to, ie profile, configuration, etc.
   */
  export function goto(event: Event, pid: string, subPages: string[] = []): void {
    event.preventDefault()
    pid = pid.toLowerCase()
    if (settings.concat(insights, others).find(p => p.id.toLowerCase() === pid))
      activePage = pid
    else activePage = ''

    const query = new URLSearchParams(window.location.search)
    const params = query.toString()
    const uri = `${$urlbase}${[pid, ...subPages].join('/')}${params ? `?${params}` : ''}`
    window.history.replaceState({}, '', uri)
    // Auto-collapse sidebar on mobile after navigation
    isOpen = windowWidth >= 992
  }

  // Used to auto-navigate.
  $: parts = ltrim(ltrim(window.location.pathname, $urlbase), '/').split('/')
  $: activePage = parts.length > 0 ? parts[0] : ''
  // Used for the language dropdown.
  $: locale = $profile.languages?.[currentLocale()]?.[currentLocale()]?.self
  $: newLang = currentLocale()

  // Used for sidebar collapse state.
  let windowWidth = 1000
  $: isOpen = windowWidth >= 992

  // Set the component based on the active page. Dig it out of settings, others and insights.
  $: PageComponent =
    settings.concat(insights, others).find(page => page.id.toLowerCase() === activePage)
      ?.component || Landing
</script>

<svelte:window bind:innerWidth={windowWidth} />

<div class="navigation">
  <!-- Mobile Menu Toggle Button -->
  <Card color="warning" {theme} class="menu-toggle-wrapper d-lg-none mb-2 p-0" outline>
    <Button size="sm" class="menu-toggle-button my-0" on:click={() => (isOpen = !isOpen)}>
      {isOpen ? `✕ ${$_('phrases.HideMenu')}` : `☰ ${$_('phrases.ShowMenu')}`}
    </Button>
  </Card>

  <Row>
    <!-- Navigation Sidebar -->
    <Fade class="sidebar-col col" {isOpen}>
      <Card body class="sidebar-card" {theme}>
        <!-- Settings -->
        <p class="navheader">{$_('navigation.titles.Settings')}</p>
        <Nav vertical pills class="nav-custom" {theme}>
          {#each settings as page}
            {@const pid = page.id.toLowerCase()}
            <NavItem>
              <NavLink
                href={$urlbase + pid}
                class="nav-link-custom"
                active={activePage === pid}
                on:click={e => goto(e, pid)}>
                <span class="nav-icon">{page.icon}</span>
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
            {@const pid = page.id.toLowerCase()}
            <NavItem>
              <NavLink
                href={$urlbase + pid}
                class="nav-link-custom"
                active={activePage === pid}
                on:click={e => goto(e, pid)}>
                <span class="nav-icon">{page.icon}</span>
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
                class="dropdown-custom {activePage === 'trustprofile' ? 'active' : ''}">
                <span class="text-uppercase profile-icon">
                  {$profile.username.charAt(0).toUpperCase()}
                </span>
                <span>{$profile.username}</span>
              </DropdownToggle>
              <DropdownMenu>
                <DropdownItem
                  class="nav-link-custom"
                  active={activePage === 'trustprofile'}
                  onclick={e => goto(e, 'trustprofile')}>
                  <span class="nav-icon">👤</span>
                  <span class="nav-text">{$_('navigation.titles.TrustProfile')}</span>
                </DropdownItem>
                {#if locale}
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
                {:else}
                  <DropdownItem divider />
                {/if}
                <DropdownItem class="nav-link-custom" onclick={toggleDarkMode}>
                  <Icon
                    name={$darkMode ? 'sun' : 'moon'}
                    class="me-3 text-{$darkMode ? 'warning' : 'primary'}" />
                  {$darkMode ? $_('config.titles.Light') : $_('config.titles.Dark')}
                  <Input
                    type="switch"
                    bind:checked={$darkMode}
                    style="position: absolute; right: 5px;" />
                </DropdownItem>
              </DropdownMenu>
            </Dropdown>
          </Nav>
        </div>
      </Card>
    </Fade>

    <!-- Content Area -->
    <Col><svelte:component this={PageComponent} /></Col>
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
    font-size: 16px;
    display: inline-flex;
    width: 20px;
    justify-content: center;
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
      transition: transform 0.3s ease;
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
