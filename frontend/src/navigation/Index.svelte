<script lang="ts" module>
  let sidebarOpen = $state(false)
  export const closeSidebar = () => (sidebarOpen = false)
  export const toggleSidebar = () => (sidebarOpen = !sidebarOpen)
</script>

<script lang="ts">
  import { Card, Col, Button } from '@sveltestrap/sveltestrap'
  import { _ } from '../includes/Translate.svelte'
  import { nav } from './nav.svelte'
  import { theme } from '../includes/theme.svelte'
  import { slide } from 'svelte/transition'
  import { onMount } from 'svelte'
  import Sidebar from './Sidebar.svelte'

  const magicNumber = 1005
  // windowWidth is used for sidebar collapse state.
  let windowWidth = $state(magicNumber + 1)
  const isMobile = $derived(windowWidth <= magicNumber)

  onMount(() => nav.onMount())
  $effect(() => {
    if (windowWidth < magicNumber) sidebarOpen = false
  })
</script>

<svelte:window bind:innerWidth={windowWidth} on:popstate={e => nav.popstate(e)} />

<div class="menu-toggle-wrapper">
  {#if isMobile}
    <!-- Mobile Menu Toggle Button -->
    <div transition:slide style="overflow: visible">
      <Card color="warning" theme={$theme} class="toggle-card mb-2 p-0" outline>
        <Button size="sm" class="toggle-button my-0" onclick={toggleSidebar}>
          {sidebarOpen ? `✕ ${$_('phrases.HideMenu')}` : `☰ ${$_('phrases.ShowMenu')}`}
        </Button>
      </Card>
    </div>
  {/if}
</div>

{#if sidebarOpen || !isMobile}
  {@const flex = isMobile ? 'flex-col' : ''}
  {@const transition = { duration: 600, axis: 'x' as const }}
  <!-- Navigation Sidebar. -->
  <div class="sidebar-col col mb-2 {flex}" transition:slide={transition}>
    <Sidebar slide={transition} />
  </div>
{/if}

<!-- Content Area -->
<Col>
  <Card class="mb-2" outline color="notifiarr" theme={$theme}>
    <nav.ActivePage />
  </Card>
</Col>

<style>
  .sidebar-col {
    min-width: 230px;
    max-width: fit-content;
    margin-right: 0px;
  }

  /* Mobile styles for menu toggler. */

  .menu-toggle-wrapper {
    position: sticky;
    top: 0px;
    z-index: 1010;
    overflow-x: visible;
  }

  .menu-toggle-wrapper :global(.toggle-card) {
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

  /* Mobile styles for sidebar. Turns it into a fixed sidebar w/ toggler. */
  .flex-col {
    position: fixed;
    z-index: 1020;
    max-height: 100vh;
    overflow-y: auto;
    top: 0;
    left: 0;
    padding: 1px;
    border-radius: 12px;
    box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
    background: rgba(118, 122, 126, 0.9);
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
