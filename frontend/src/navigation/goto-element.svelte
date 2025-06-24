<!-- Used in translations to link to a local page. -->
<svelte:options customElement={{ tag: 'go-to', shadow: 'open' }} />

<script>
  import { urlbase } from '../api/fetch'
  import { nav } from './nav.svelte'
  import { theme } from '../includes/theme.svelte'
  let { page = 'landing' } = $props()
  const darkMode = $derived($theme.includes('dark'))
</script>

<a href={$urlbase + page} onclick={e => nav.goto(e, page)} class:darkMode>
  <slot />
</a>

<style>
  /* These match the 'a' colors in the app.css file. */

  a {
    color: #2fa582 !important;
  }

  a:hover {
    color: #3cd2a5 !important;
  }

  a.darkMode {
    color: #3cd2a5 !important;
  }

  a.darkMode:hover {
    color: #2fa582 !important;
  }
</style>
