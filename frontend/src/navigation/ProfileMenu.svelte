<script lang="ts">
  import {
    Dropdown,
    DropdownItem,
    DropdownMenu,
    DropdownToggle,
    Nav,
  } from '@sveltestrap/sveltestrap'
  import { nav } from './nav.svelte'
  import { theme } from '../includes/theme.svelte'
  import { _ } from '../includes/Translate.svelte'
  import { currentLocale, setLocale, Flags } from '../includes/locale/index.svelte'
  import { profile } from '../api/profile.svelte'
  import Fa from '../includes/Fa.svelte'
  import { Input } from '@sveltestrap/sveltestrap'
  import { faStarship, faSun } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { page as ProfilePage } from '../pages/profile/Index.svelte'

  let newLang = $derived(currentLocale())
</script>

<Nav vertical pills class="nav-custom" theme={$theme}>
  <Dropdown nav direction="up" class="ms-0">
    <DropdownToggle
      class="dropdown-custom {nav.active(ProfilePage.id) ? 'active' : ''}"
      nav>
      <span class="text-uppercase profile-icon">
        {$profile.username.charAt(0).toUpperCase()}
      </span>
      {$profile.username}
    </DropdownToggle>
    <DropdownMenu>
      <DropdownItem
        class="nav-link-custom"
        active={nav.active(ProfilePage.id)}
        disabled={nav.active(ProfilePage.id)}
        onclick={e => nav.goto(e, ProfilePage.id)}>
        <span class="nav-icon"><Fa {...ProfilePage} scale="1.7x" /></span>
        {$_('navigation.titles.TrustProfile')}
      </DropdownItem>
      <span class="lang-wrapper">
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
      </span>
      <DropdownItem class="nav-link-custom" onclick={theme.toggle}>
        <Fa
          i={faStarship}
          d={faSun}
          c1="green"
          c2="lightblue"
          d1="orange"
          d2="fuchsia"
          scale="1.3x"
          class="me-3" />
        {theme.isDark ? $_('config.titles.Light') : $_('config.titles.Dark')}
        <Input
          disabled
          type="switch"
          checked={theme.isDark}
          style="position: absolute; right: 5px;" />
      </DropdownItem>
    </DropdownMenu>
  </Dropdown>
</Nav>

<style>
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

  .lang-wrapper :global(.lang-select) {
    padding: 1px 0px 1px 14px;
    border: 0px;
    font-size: 16px;
  }
</style>
