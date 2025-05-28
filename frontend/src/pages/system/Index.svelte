<script lang="ts" module>
  import { faServer } from '@fortawesome/sharp-duotone-regular-svg-icons'
  export const page = {
    id: 'System',
    i: faServer,
    c1: 'darkslategray',
    c2: 'dimgray',
    d1: 'lavender',
    d2: 'gainsboro',
  }
</script>

<script lang="ts">
  import { CardBody, Row, Col } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import Header from '../../includes/Header.svelte'
  import Environment from './Environment.svelte'
  import Storage from './Storage.svelte'
  import Network from './Network.svelte'
  import StartupParams from './StartupParams.svelte'
  import BuildInfo from './BuildInfo.svelte'
  import ClientInfo from './ClientInfo.svelte'
  import Os from './OS.svelte'
  import UserInfo from './UserInfo.svelte'
</script>

<Header {page} />

<div class="system-wrapper">
  <CardBody class="pt-0 mt-0">
    <Row>
      <Col md={6}><UserInfo /></Col>
      <Col md={6}><Os /></Col>
    </Row>
    <Row>
      <Col md={6}><ClientInfo /></Col>
      <Col md={6}><BuildInfo /></Col>
    </Row>
    <Row>
      <Col md={6}><StartupParams /></Col>
      <Col md={6}><Network /></Col>
    </Row>
    <Row><Col md={12}><Storage /></Col></Row>
    <Row><Col md={12}><Environment /></Col></Row>
  </CardBody>
</div>

<style>
  .system-wrapper :global(td) {
    word-break: break-all;
  }

  .system-wrapper :global(tr) {
    line-height: 1.1;
  }
</style>
