<!--
  This is the main navigation component.
  It is responsible for displaying the sidebar and the content area.
  This component, with the help of nav.svelte.ts keeps the window address bar accurate.
  It also handles the sidebar toggle for mobile devices.
-->
<script lang="ts" module>
  let sidebarOpen = $state(false)
  export const closeSidebar = () => (sidebarOpen = false)
  export const toggleSidebar = () => (sidebarOpen = !sidebarOpen)
</script>

<script lang="ts">
  import { Card, Col, Button } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../includes/Translate.svelte'
  import { nav } from './nav.svelte'
  import { theme } from '../includes/theme.svelte'
  import { fade, slide } from 'svelte/transition'
  import { onMount, tick } from 'svelte'
  import Sidebar from './Sidebar.svelte'
  import Modals from './Modals.svelte'
  import Nodal from '../includes/Nodal.svelte'

  const magicNumber = 1005
  // windowWidth is used for sidebar collapse state. Default desktop so the first
  // paint does not stack a mobile drawer above the content card.
  let windowWidth = $state(
    typeof window !== 'undefined' ? window.innerWidth : magicNumber + 1,
  )
  const isMobile = $derived(windowWidth <= magicNumber)
  // Desktop: stretch the sticky sidebar to the content height.
  let contentHeight = $state(0)
  // First page paints immediately; later navigations fade in the same slot.
  let pageFade = $state(false)

  onMount(async () => {
    await tick()
    await nav.onMount()
    pageFade = true
  })

  // Close the overlay only when crossing desktop → mobile, not on every
  // innerWidth tick (rotation would otherwise dismiss an open menu).
  const setWindowWidth = (width: number) => {
    if (width <= magicNumber && windowWidth > magicNumber) sidebarOpen = false
    windowWidth = width
  }
</script>

<svelte:window
  bind:innerWidth={() => windowWidth, setWindowWidth}
  on:popstate={e => nav.popstate(e)} />

<Modals />

{#if isMobile}
  <div class="menu-toggle-wrapper">
    <!-- Mobile Menu Toggle Button -->
    <div transition:slide style="overflow: visible">
      <Card color="warning" theme={$theme} class="toggle-card mb-2 p-0" outline>
        <Button size="sm" class="toggle-button my-0" onclick={toggleSidebar}>
          {#if sidebarOpen}
            <T id="buttons.HideMenu" />
          {:else}
            <T id="buttons.ShowMenu" />
          {/if}
        </Button>
      </Card>
    </div>
  </div>
{/if}

{#if sidebarOpen || !isMobile}
  {@const flex = isMobile ? 'flex-col' : ''}
  {@const transition = { duration: 600, axis: 'x' as const }}
  <!-- Navigation Sidebar. -->
  <div class="sidebar-col col mb-2 {flex}" transition:slide={transition}>
    <Sidebar slide={transition} height={contentHeight} {isMobile} />
  </div>
{/if}

<!-- Content Area -->
<Col style="width: 1%;">
  <Card class="mb-2" outline color="notifiarr" theme={$theme}>
    <div class="page-stack" bind:clientHeight={contentHeight}>
      {#key nav.ActivePage}
        <div
          class="page-layer"
          in:fade={pageFade ? { duration: 180 } : { duration: 0 }}
          out:fade={{ duration: 180 }}>
          <nav.ActivePage />
        </div>
      {/key}
    </div>
  </Card>
</Col>

<!-- This uses global variables to show a modal whenever any (connected)
     form has changes and you might lose them by navigating away. -->
<Nodal
  isOpen={nav.showUnsavedAlert !== null}
  title="navigation.titles.UnsavedChanges"
  follow={() => (nav.showUnsavedAlert = null)}
  esc>
  <T id="phrases.LeavePage" />
  {#snippet footer()}
    <Button color="primary" onclick={() => (nav.showUnsavedAlert = null)}>
      <T id="buttons.NoStayHere" />
    </Button>
    <Button
      color="danger"
      onclick={() => nav.goto(nav.forceEvent, nav.showUnsavedAlert ?? '')}>
      <T id="buttons.YesDeleteMyChanges" />
    </Button>
  {/snippet}
</Nodal>

<style>
  .sidebar-col {
    min-width: 170px;
    max-width: fit-content;
  }

  /* Outgoing and incoming pages share one cell so they fade instead of stacking. */
  .page-stack {
    display: grid;
  }

  .page-layer {
    grid-area: 1 / 1;
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

  /* Mobile drawer: fill the visual viewport and let Sidebar scroll internally. */
  .flex-col {
    position: fixed;
    z-index: 1020;
    top: 0;
    left: 0;
    bottom: 0;
    height: 100dvh;
    max-height: 100dvh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    padding: 1px;
    border-radius: 12px;
    box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
    background: rgba(118, 122, 126, 0.9);
  }

  .flex-col :global(.sidebar-card-wrapper) {
    flex: 1 1 auto;
    min-height: 0;
    height: 100%;
  }
</style>
