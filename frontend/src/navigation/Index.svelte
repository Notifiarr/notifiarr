<script lang="ts">
  import { Card, Row, Col, Button, Theme } from '@sveltestrap/sveltestrap'
  import { _ } from '../includes/Translate.svelte'
  import { nav } from './nav.svelte'
  import { settings, insights } from './pages'
  import { theme } from '../includes/theme.svelte'
  import { slide } from 'svelte/transition'
  import { onMount } from 'svelte'
  import ProfileMenu from './ProfileMenu.svelte'
  import NavSection from './NavSection.svelte'

  // windowWidth is used for sidebar collapse state.
  let windowWidth = $state(1000)
  let isOpen = $derived(windowWidth >= 992)
  onMount(() => nav.onMount())
</script>

<svelte:window bind:innerWidth={windowWidth} on:popstate={e => nav.popstate(e)} />

<div class="menu-toggle-wrapper">
  <!-- Mobile Menu Toggle Button -->
  <Card color="warning" theme={$theme} class="toggle-card d-lg-none mb-2 p-0" outline>
    <Button size="sm" class="toggle-button my-0" onclick={() => (isOpen = !isOpen)}>
      {isOpen ? `✕ ${$_('phrases.HideMenu')}` : `☰ ${$_('phrases.ShowMenu')}`}
    </Button>
  </Card>
</div>

<Row>
  <!-- Navigation Sidebar -->
  {#if isOpen}
    <div class="sidebar-col col mb-2" transition:slide={{ duration: 475, axis: 'x' }}>
      <Card body class="sidebar-card pb-2" theme={$theme}>
        <!-- Settings -->
        <NavSection title="Settings" pages={settings} />
        <div class="section-divider"></div>
        <!-- Insights -->
        <NavSection title="Insights" pages={insights} />
        <!-- Profile Dropdown -->
        <ProfileMenu />
      </Card>
    </div>
  {/if}

  <!-- Content Area -->
  <Col>
    <Theme theme={$theme}>
      <Card class="mb-2" outline color="notifiarr">
        <nav.ActivePage />
      </Card>
    </Theme>
  </Col>
</Row>

<style>
  /* Mobile styles */
  .menu-toggle-wrapper :global(.toggle-card) {
    position: sticky;
    top: 0;
    z-index: 1010;
    padding: 5px;
    margin-bottom: 15px;
    text-align: center;
    border-radius: 3px;
  }

  .menu-toggle-wrapper :global(.toggle-button) {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    width: 100%;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  }

  .section-divider {
    height: 1px;
    background-color: rgba(0, 0, 0, 0.1);
    margin: 12px 0;
    width: 100%;
  }

  .sidebar-col {
    min-width: 230px;
    max-width: 230px;
  }

  .sidebar-col :global(.sidebar-card) {
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

  @media (max-width: 991.98px) {
    .sidebar-col {
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

  /* These styles are used by ProfileMenu.svelte and NavSection.svelte */

  .sidebar-col :global(.nav-custom) {
    gap: 4px;
  }

  .sidebar-col :global(.nav-link-custom) {
    display: flex;
    align-items: center;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .sidebar-col :global(.nav-link-custom.active) {
    background: linear-gradient(135deg, #1a73e8, #6c5ce7);
    box-shadow: 0 2px 5px rgba(108, 92, 231, 0.2);
  }

  .sidebar-col :global(.nav-icon) {
    margin-right: 12px;
    display: inline-flex;
    width: 20px;
  }

  /* This is only in ProfileMenu.svelte, but shares a gradient with the above. */

  .sidebar-col :global(.dropdown-custom.active) {
    background: linear-gradient(135deg, #1a73e8, #6c5ce7);
    box-shadow: 0 2px 5px rgba(108, 92, 231, 0.2);
    transition: all 0.2s ease;
    border-radius: 8px;
    color: #f1f3f5;
  }
</style>
