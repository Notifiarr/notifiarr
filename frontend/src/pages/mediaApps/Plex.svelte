<!-- this is a standalone component page for the `/plex` URI -->
<script lang="ts" module>
  import { faLocationQuestion } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { CardBody, Col } from '@sveltestrap/sveltestrap'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import Plex from '../integrations/Plex.svelte'

  export const page = {
    id: 'Plex',
    i: faLocationQuestion,
    c1: 'blue',
    c2: 'wheat',
    d1: 'darkcyan',
    d2: 'gold',
  }
</script>

<Header {page} />

<CardBody>
  {#if $profile.config.plex.url == ''}
    <ul><li class="text-danger"><T id="Plex.URLNotConfigured" /></li></ul>
  {:else if $profile.config.plex.token == ''}
    <ul><li class="text-danger"><T id="Plex.TokenNotConfigured" /></li></ul>
  {:else if $profile.config.plex.timeout == '-1s'}
    <ul><li class="text-danger"><T id="Plex.Disabled" /></li></ul>
  {:else}
    {@const count = $profile.expvar.apps?.Plex?.['Incoming Webhooks'] ?? 0}
    <ul>
      <li class="text-success"><T id="Plex.WaitingForWebhooks" /></li>
      <li><T id="Plex.WebhooksReceived" {count} /></li>
      {#if !$profile.plexInfo?.friendlyName}
        <li class="text-danger"><T id="Plex.NoStatus" /></li>
      {/if}
    </ul>
  {/if}

  {#if $profile.plexInfo?.friendlyName}
    <Col sm={10} md={8} lg={7} xl={6} xxl={5}>
      <Plex
        status={$profile.plexInfo}
        plexAge={$profile.plexAge}
        showSessions={false}
        showOwner={false} />
    </Col>
  {/if}
</CardBody>
