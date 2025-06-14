<!-- Used in translations to link to a local page. -->
<svelte:options customElement={{ tag: 'goto-page', shadow: 'open' }} />

<script>
  import { urlbase } from '../api/fetch'
  import { nav } from './nav.svelte'
  let { page = 'landing' } = $props()
</script>

<a href={$urlbase + page} onclick={e => nav.goto(e, page)}><slot /></a>
