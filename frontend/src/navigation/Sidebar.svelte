<script lang="ts">
  import { Card } from '@sveltestrap/sveltestrap'
  import Section from './Section.svelte'
  import { theme } from '../includes/theme.svelte'
  import { settings, insights } from './pages'
  import ProfileMenu from './ProfileMenu.svelte'
  import { slide as slyde } from 'svelte/transition'

  // This is how many pixels we need to display the entire sidebar.
  // If you add elements, you may need to increase this.
  const h: number = 900
  // The slide parameters have to match the parent element, so it's passed in.
  let { slide = { duration: 600, axis: 'x' }, height = h } = $props()
  // This is a hack to get the sidebar to expand up to the height of the content.
  const style = $derived(`max-height: ${height > h ? height + 1 : h}px !important;`)
</script>

<div class="sidebar-card-wrapper" transition:slyde={slide} {style}>
  <Card body class="sidebar-card pb-2" theme={$theme} {style}>
    <!-- Settings -->
    <Section title="Settings" pages={settings} />
    <div class="section-divider"></div>
    <!-- Insights -->
    <Section title="Insights" pages={insights} />
    <!-- Profile Dropdown -->
    <div class="mt-auto pt-2">
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

  /* These styles are used by ProfileMenu.svelte and Section.svelte */

  .sidebar-card-wrapper :global(.nav-custom) {
    gap: 4px;
  }

  .sidebar-card-wrapper :global(.nav-link-custom) {
    display: flex;
    align-items: center;
    border-radius: 8px;
    transition: all 0.2s ease;
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

  .sidebar-card-wrapper :global(.nav-icon) {
    margin-right: 12px;
    display: inline-flex;
    width: 20px;
  }
</style>
