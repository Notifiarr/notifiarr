<script lang="ts">
  import { Card } from '@sveltestrap/sveltestrap'
  import Section from './Section.svelte'
  import { theme } from '../includes/theme.svelte'
  import { settings, insights } from './pages'
  import ProfileMenu from './ProfileMenu.svelte'
  import { slide as slyde } from 'svelte/transition'

  // Desktop only: stretch the sidebar card up to the content height so a
  // tall Config page does not leave a short sticky nav. Mobile ignores this
  // and fills the viewport drawer instead.
  const h: number = 900
  let { slide = { duration: 600, axis: 'x' }, height = h, isMobile = false } = $props()
  const style = $derived(
    isMobile ? '' : `max-height: ${height > h ? height + 1 : h}px !important;`,
  )
</script>

<div class="sidebar-card-wrapper" class:mobile={isMobile} transition:slyde={slide} {style}>
  <Card body class="sidebar-card pb-2" theme={$theme}>
    <div class="sidebar-scroll">
      <Section title="Settings" pages={settings} />
      <div class="section-divider"></div>
      <Section title="Insights" pages={insights} />
    </div>
    <div class="sidebar-profile mt-auto pt-2">
      <div class="section-divider"></div>
      <ProfileMenu />
    </div>
  </Card>
</div>

<style>
  .sidebar-card-wrapper {
    position: sticky;
    top: 10px;
    display: flex;
    flex-direction: column;
    overflow-y: visible;
    overflow-x: visible;
    min-height: calc(100vh - 150px);
  }

  .sidebar-card-wrapper.mobile {
    position: static;
    top: auto;
    min-height: 0;
    height: 100%;
    overflow: hidden;
  }

  .section-divider {
    height: 2px;
    background-color: var(--bs-secondary-bg-subtle);
    margin: 12px 0;
    width: 100%;
  }

  .sidebar-card-wrapper :global(.sidebar-card) {
    border-radius: 12px;
    padding: 10px 5px 10px 5px;
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.05);
  }

  .sidebar-card-wrapper.mobile :global(.sidebar-card) {
    display: flex;
    flex-direction: column;
    flex: 1 1 auto;
    height: 100%;
    min-height: 0;
    overflow: hidden;
  }

  .sidebar-card-wrapper.mobile .sidebar-scroll {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    overscroll-behavior: contain;
  }

  .sidebar-profile {
    flex-shrink: 0;
  }

  /* These styles are used by ProfileMenu.svelte and Section.svelte */

  .sidebar-card-wrapper :global(.nav-custom) {
    gap: 4px;
  }

  .sidebar-card-wrapper :global(.nav-link-custom) {
    display: flex;
    align-items: center;
    border-radius: 8px;
    transition: all 0.2s ease;
    padding-top: 5px;
    padding-bottom: 5px;
  }

  /* Make all the sidebar selectors look really nice. */
  .sidebar-card-wrapper :global(.nav-link-custom.active),
  .sidebar-card-wrapper :global(.nav-link-custom:hover),
  .sidebar-card-wrapper :global(.dropdown-custom:hover),
  .sidebar-card-wrapper :global(.dropdown-custom.show),
  .sidebar-card-wrapper :global(.dropdown-custom.active),
  .sidebar-card-wrapper :global(select:hover) {
    background: linear-gradient(135deg, rgb(81, 191, 158), #359fa4);
    box-shadow: 0 2px 5px rgba(92, 231, 201, 0.454) !important;
    transition: all 0.5s ease !important;
    border-radius: 8px;
    color: ivory !important;
  }

</style>
