<!-- this is a standalone component page for the `/plex` URI -->
<script lang="ts" module>
  import { faLocationQuestion } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { Card, CardBody, CardHeader, Col, Row, Table } from '@sveltestrap/sveltestrap'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import Plex from '../integrations/Plex.svelte'
  import { urlbase } from '../../api/fetch'
  import { fade } from 'svelte/transition'

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
  <Row>
    {#if $profile.plexInfo?.friendlyName}
      <Col md={6} xxl={4} class="mb-2">
        <Plex
          status={$profile.plexInfo}
          plexAge={$profile.plexAge}
          showSessions={false}
          showOwner={false} />
      </Col>
    {/if}
    {#if $profile.clientInfo?.actions.plex && $profile.plexInfo?.friendlyName}
      <Col md={6} xxl={4} class="mb-2">
        <Card color="warning" outline>
          <CardHeader tag="div"><h5><T id="Plex.WebsiteSettings" /></h5></CardHeader>
          <Table class="rounded-bottom mb-0" size="sm">
            <tbody>
              <tr>
                <td><T id="Plex.AccountMap" /></td>
                <td>{$profile.clientInfo?.actions.plex.accountMap}</td>
              </tr>
              <tr>
                <td><T id="Plex.ActivityDelay" /></td>
                <td>{$profile.clientInfo?.actions.plex.activityDelay}</td>
              </tr>
              <tr>
                <td><T id="Plex.WebhookCoolDown" /></td>
                <td>{$profile.clientInfo?.actions.plex.cooldown}</td>
              </tr>
              <tr>
                <td><T id="Plex.SessionsInterval" /></td>
                <td>{$profile.clientInfo?.actions.plex.interval}</td>
              </tr>
              <tr>
                <td><T id="Plex.MoviesFinished" /></td>
                <td>{$profile.clientInfo?.actions.plex.moviesPc}%</td>
              </tr>
              <tr>
                <td><T id="Plex.SeriesFinished" /></td>
                <td>{$profile.clientInfo?.actions.plex.seriesPc}%</td>
              </tr>
              <tr>
                <td><T id="Plex.NoActivity" /></td>
                <td>{$profile.clientInfo?.actions.plex.noActivity}</td>
              </tr>
            </tbody>
          </Table>
        </Card>
      </Col>
    {/if}
  </Row>
  <Row>
    <Col>
      <T
        id="Plex.WebhookNote"
        webhookUrl={window.location.origin +
          $urlbase +
          'plex?token=' +
          $profile.config.plex.token}
        urlbase={$urlbase} />
    </Col>
  </Row>
</CardBody>

<style>
  :global(table tbody tr td:first-child) {
    white-space: nowrap;
    width: 1%;
    padding-right: 0.7rem;
  }
</style>
