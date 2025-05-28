<script lang="ts">
  import { Nav, NavItem, NavLink } from '@sveltestrap/sveltestrap'
  import { nav, type Page } from './nav.svelte'
  import Fa from '../includes/Fa.svelte'
  import { urlbase } from '../api/fetch'
  import { _ } from '../includes/Translate.svelte'
  import { theme } from '../includes/theme.svelte'

  type Props = { title: string; pages: Page[] }

  const { title, pages }: Props = $props()
</script>

<p class="nav-header">{$_('navigation.titles.' + title)}</p>
<div class="nav-section">
  <Nav vertical pills class="nav-custom" theme={$theme}>
    {#each pages as page}
      <NavItem>
        <NavLink
          href={$urlbase + page.id}
          class="nav-link-custom"
          active={nav.active(page.id)}
          disabled={nav.active(page.id)}
          onclick={e => nav.goto(e, page.id)}>
          <span class="nav-icon">
            <Fa {...page} scale="1.7x" />
          </span>
          <span class="nav-text">{$_('navigation.titles.' + page.id)}</span>
        </NavLink>
      </NavItem>
    {/each}
  </Nav>
</div>

<style>
  .nav-header {
    font-weight: 600;
    font-size: 12px;
    text-transform: uppercase;
    color: lightslategray;
    margin-top: 5px;
    margin-bottom: 8px;
    padding-left: 8px;
    letter-spacing: 0.5px;
  }
</style>
