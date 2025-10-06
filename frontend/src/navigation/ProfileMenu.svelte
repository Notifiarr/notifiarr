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
  import { locale, Flags } from '../includes/locale/locale.svelte'
  import { profile } from '../api/profile.svelte'
  import Fa from '../includes/Fa.svelte'
  import { Input } from '@sveltestrap/sveltestrap'
  import { faStarship, faSun } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { page as ProfilePage } from '../pages/profile/Index.svelte'
  import { page as LanguagesPage } from '../pages/stubs/Languages.svelte'
  import { closeSidebar } from './Index.svelte'

  let newLang = $derived(locale.current)
  const onchange = () => (locale.set(newLang), closeSidebar())
</script>

<Nav vertical pills class="nav-custom" theme={$theme}>
  <Dropdown nav direction="up" class="ms-0" setActiveFromChild>
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
        <span class="nav-icon"><Fa {...ProfilePage} scale="1.2x" /></span>
        {$_('navigation.titles.' + ProfilePage.id)}
      </DropdownItem>
      <span class="lang-wrapper">
        <Input type="select" bind:value={newLang} {onchange}>
          {#each Object.entries($profile.languages?.[locale.current] || {}) as [code, lang]}
            <option value={code} selected={code === locale.current}>
              {Flags[code]}&nbsp;&nbsp; {lang.name}
            </option>
          {/each}
        </Input>
      </span>
      <DropdownItem class="nav-link-custom" onclick={e => nav.goto(e, LanguagesPage.id)}>
        <span class="nav-icon"><Fa {...LanguagesPage} scale="1.2x" /></span>
        {$_(LanguagesPage.id + '.menuTitle')}
      </DropdownItem>
      <DropdownItem class="nav-link-custom" onclick={theme.toggle}>
        <Fa
          i={faStarship}
          d={faSun}
          c1="green"
          c2="lightblue"
          d1="orange"
          d2="fuchsia"
          scale="1.2x"
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

  .lang-wrapper :global(select) {
    padding: 1px 0px 1px 14px;
    margin: 5px 0px 5px 0px;
    border: 0px;
    font-size: 16px;
    height: 32px;
    min-width: 180px;
  }
</style>
