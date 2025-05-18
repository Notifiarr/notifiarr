<script lang="ts">
  import { Card } from '@sveltestrap/sveltestrap'
  import Section from './Section.svelte'
  import { theme } from '../includes/theme.svelte'
  import { settings, insights } from './pages'
  import ProfileMenu from './ProfileMenu.svelte'
  import { slide as sly } from 'svelte/transition'

  let { slide = { duration: 600, axis: 'x' } } = $props()
</script>

<div class="sidebar-card-wrapper" transition:sly={slide}>
  <Card body class="sidebar-card pb-2" theme={$theme}>
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
    overflow-y: auto;
    overflow-x: visible;
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
    min-height: calc(100vh - 150px);
  }
</style>
