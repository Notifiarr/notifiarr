<script lang="ts">
  import { Row, Col } from '@sveltestrap/sveltestrap'
  import Input from '../../includes/Input.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import type { Config } from '../../api/notifiarrConfig'

  type Props = { config: Config; original: Config }
  const { config = $bindable(), original }: Props = $props()
</script>

<!-- SSL Section -->
<h4>{$_('config.titles.SSLConfiguration')}</h4>
<Row>
  <Col md={6}>
    <Input
      id="config.sslKeyFile"
      envVar="SSL_KEY_FILE"
      bind:value={config.sslKeyFile}
      original={original.sslKeyFile} />
  </Col>
  <Col md={6}>
    <Input
      id="config.sslCertFile"
      envVar="SSL_CERT_FILE"
      bind:value={config.sslCertFile}
      original={original.sslCertFile} />
  </Col>
</Row>
