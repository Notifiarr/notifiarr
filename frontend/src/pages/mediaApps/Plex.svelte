<!-- this is a standalone component page for the `/plex` URI -->
<script lang="ts" module>
  import { faLocationQuestion } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import T, { _ } from '../../includes/Translate.svelte'

  export const page = {
    id: 'Plex',
    i: faLocationQuestion,
    c1: 'blue',
    c2: 'wheat',
    d1: 'purple',
    d2: 'goldenrod',
  }
</script>

<Header {page} />

<CardBody>
  {#if $profile.config.plex.url == ''}
    <T id="Plex.URLNotConfigured"></T>
  {:else if $profile.config.plex.token == ''}
    <T id="Plex.TokenNotConfigured"></T>
  {:else if $profile.config.plex.timeout == '-1s'}
    <T id="Plex.Disabled"></T>
  {:else}
    {@const count = $profile.expvar.apps?.Plex?.['Incoming Webhooks'] ?? 0}
    <T id="Plex.WaitingForWebhooks" {count} />
  {/if}
</CardBody>
