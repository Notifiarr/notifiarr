<script lang="ts">
  import { Row, Col } from '@sveltestrap/sveltestrap'
  import Input from '../../includes/Input.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import type { Config } from '../../api/notifiarrConfig'
  import BrowserModal from '../../includes/fileBrowser/BModal.svelte'
  import BButton from '../../includes/fileBrowser/BButton.svelte'

  type Props = { config: Config; original: Config }
  const { config = $bindable(), original }: Props = $props()
  let keyModal = $state(false)
  let certModal = $state(false)
</script>

<!-- SSL Section -->
<h4>{$_('config.titles.SSLConfiguration')}</h4>
<Row>
  <Col md={6}>
    <Input
      id="config.sslKeyFile"
      bind:value={config.sslKeyFile}
      original={original.sslKeyFile}>
      {#snippet post()}<BButton bind:isOpen={keyModal} />{/snippet}
    </Input>
  </Col>
  <Col md={6}>
    <Input
      id="config.sslCertFile"
      bind:value={config.sslCertFile}
      original={original.sslCertFile}>
      {#snippet post()}<BButton bind:isOpen={certModal} />{/snippet}
    </Input>
  </Col>
</Row>

<BrowserModal
  file
  bind:isOpen={keyModal}
  bind:value={config.sslKeyFile}
  title="config.sslKeyFile.label">
  <p><T id="config.sslKeyFile.description" /></p>
  <T id="config.sslKeyFile.tooltip" />
</BrowserModal>

<BrowserModal
  file
  bind:isOpen={certModal}
  bind:value={config.sslCertFile}
  title="config.sslCertFile.label">
  <p><T id="config.sslCertFile.description" /></p>
  <T id="config.sslCertFile.tooltip" />
</BrowserModal>
