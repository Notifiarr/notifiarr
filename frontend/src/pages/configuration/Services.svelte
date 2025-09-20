<script lang="ts">
  import { Row, Col } from '@sveltestrap/sveltestrap'
  import Input from '../../includes/Input.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import type { Config } from '../../api/notifiarrConfig'

  type Props = { config: Config; original: Config }
  const { config = $bindable(), original }: Props = $props()
</script>

<!-- Services Section -->
<h4>{$_('config.titles.Services')}</h4>
<Row>
  <Col md={4}>
    <Input
      id="config.services.disabled"
      type="select"
      bind:value={config.services!.disabled}
      original={original.services?.disabled} />
  </Col>
  <Col md={4}>
    <Input
      id="config.services.parallel"
      type="select"
      options={[
        { value: 1, name: '1' },
        { value: 2, name: '2' },
        { value: 3, name: '3' },
        { value: 4, name: '4' },
        { value: 5, name: '5' },
      ]}
      bind:value={config.services!.parallel}
      original={original.services?.parallel} />
  </Col>
  <Col md={4}>
    <Input
      id="config.services.interval"
      type="select"
      bind:value={config.services!.interval}
      options={[
        { value: '5m0s', name: '5 ' + $_('words.select-option.minutes') },
        { value: '10m0s', name: '10 ' + $_('words.select-option.minutes') },
        { value: '15m0s', name: '15 ' + $_('words.select-option.minutes') },
        { value: '20m0s', name: '20 ' + $_('words.select-option.minutes') },
        { value: '30m0s', name: '30 ' + $_('words.select-option.minutes') },
      ]}
      original={original.services?.interval} />
  </Col>
</Row>
