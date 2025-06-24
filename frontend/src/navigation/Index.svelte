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
  import {
    Card,
    Col,
    Button,
    Modal,
    ModalHeader,
    ModalBody,
    ModalFooter,
  } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../includes/Translate.svelte'
  import { nav } from './nav.svelte'
  import { theme } from '../includes/theme.svelte'
  import { slide } from 'svelte/transition'
  import { onMount } from 'svelte'
  import Sidebar from './Sidebar.svelte'
  import Modals from './Modals.svelte'

  const magicNumber = 1005
  // windowWidth is used for sidebar collapse state.
  let windowWidth = $state(magicNumber - 1)
  const isMobile = $derived(windowWidth <= magicNumber)
  // Use this to limit the sidebar height to the content height.
  let contentHeight = $state(0)

  onMount(() => nav.onMount())
  $effect(() => {
    if (windowWidth < magicNumber) sidebarOpen = false
  })
</script>

<svelte:window bind:innerWidth={windowWidth} on:popstate={e => nav.popstate(e)} />

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
    <Sidebar slide={transition} height={contentHeight} />
  </div>
{/if}

<!-- Content Area -->
<Col style="width: 1%;">
  <Card class="mb-2" outline color="notifiarr" theme={$theme}>
    {#key nav.ActivePage}
      <div bind:clientHeight={contentHeight} transition:slide>
        <nav.ActivePage />
      </div>
    {/key}
  </Card>

  <!-- This uses global variables to show a modal whenever any (connected)
     form has changes and you might lose them by navigating away. -->
  <Modal
    isOpen={nav.showUnsavedAlert !== ''}
    theme={$theme}
    color="warning"
    zIndex={9999}>
    <ModalHeader><h5><T id="navigation.titles.UnsavedChanges" /></h5></ModalHeader>
    <ModalBody><T id="phrases.LeavePage" /></ModalBody>
    <ModalFooter>
      <Button color="primary" onclick={() => (nav.showUnsavedAlert = '')}>
        <T id="buttons.NoStayHere" />
      </Button>
      <Button
        color="danger"
        onclick={() => nav.goto(nav.forceEvent, nav.showUnsavedAlert)}>
        <T id="buttons.YesDeleteMyChanges" />
      </Button>
    </ModalFooter>
  </Modal>
</Col>

<style>
  .sidebar-col {
    min-width: 230px;
    max-width: fit-content;
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
</style>
