<script lang="ts" module>
  import { faChartLine } from '@fortawesome/sharp-duotone-regular-svg-icons'
  export const page = {
    id: 'Metrics',
    i: faChartLine,
    c1: 'darkgoldenrod',
    c2: 'darkorange',
    d1: 'darkkhaki',
    d2: 'peachpuff',
  }
</script>

<script lang="ts">
  import { CardBody, Row, Col } from '@sveltestrap/sveltestrap'
  import T from '../../includes/Translate.svelte'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import { mapLength, age, warning } from '../../includes/util'
  import Expvar from './Expvar.svelte'
  import Expcard from './Expcard.svelte'

  let refresh = $state(false)
  const onclick = async (e: Event) => {
    refresh = true
    e.preventDefault()
    try {
      await profile.refresh()
    } catch (error) {
      warning(`${error}`)
    } finally {
      refresh = false
    }
  }

  const components = $derived(
    [
      { id: 'logFiles', data: $profile.expvar.logFiles },
      { id: 'apiHits', data: $profile.expvar.apiHits },
      { id: 'httpRequests', data: $profile.expvar.httpRequests },
      { id: 'website', data: $profile.expvar.website },
      { id: 'fileWatcher', data: $profile.expvar.fileWatcher },
    ].sort((a, b) => mapLength(b.data) - mapLength(a.data)),
  )
</script>

<Header {page} />
<CardBody>
  <T id="metrics.CountersReset" /><br />
  <T
    id="metrics.ApplicationUptime"
    timeDuration={age(profile.now - new Date($profile.started).getTime(), true)} />
  &nbsp;
  {#if refresh}
    <T id="phrases.UpdatingBackEnd" />
  {:else}
    <a href="#refresh_backend" {onclick}><T id="phrases.UpdateBackend" /></a>
  {/if}

  <Row>
    {#each components as { id, data }}
      <Col md={6} xxl={4}><Expvar {id} {data} /></Col>
    {/each}
    <Col md={12}><Expcard id="apps" data={$profile.expvar.apps} /></Col>
    <Col md={12}><Expcard id="timerEvents" data={$profile.expvar.timerEvents} /></Col>
  </Row>
</CardBody>
