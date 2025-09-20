<script lang="ts">
  import { Row, Col } from '@sveltestrap/sveltestrap'
  import Input from '../../includes/Input.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import type { Config } from '../../api/notifiarrConfig'

  type Props = { config: Config; original: Config }
  const { config = $bindable(), original }: Props = $props()
</script>

<!-- Network Section -->
<h4>{$_('config.titles.Network')}</h4>
<Row>
  <Col md={6}>
    <Input
      id="config.bindAddr"
      bind:value={config.bindAddr}
      original={original.bindAddr} />
  </Col>
  <Col md={6}>
    <Input id="config.urlbase" bind:value={config.urlbase} original={original.urlbase} />
  </Col>
</Row>

<Row>
  <Col md={6}>
    <Input
      id="config.timeout"
      type="select"
      bind:value={config.timeout}
      options={[
        { value: '15s', name: '15 ' + $_('words.select-option.seconds') },
        { value: '30s', name: '30 ' + $_('words.select-option.seconds') },
        { value: '1m0s', name: '1 ' + $_('words.select-option.minute') },
        { value: '2m0s', name: '2 ' + $_('words.select-option.minutes') },
        { value: '3m0s', name: '3 ' + $_('words.select-option.minutes') },
      ]}
      original={original.timeout} />
  </Col>
  <Col md={6}>
    <Input
      id="config.retries"
      type="number"
      bind:value={config.retries}
      min={0}
      original={original.retries} />
  </Col>
</Row>
