<script lang="ts">
  import { Row, Col, InputGroupText } from '@sveltestrap/sveltestrap'
  import Input from '../../includes/Input.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import type { Config } from '../../api/notifiarrConfig'

  type Props = { config: Config; original: Config }
  const { config = $bindable(), original }: Props = $props()
</script>

<!-- Logging Section -->
<h4>{$_('config.titles.Logging')}</h4>
<Row>
  <Col md={6}>
    <Input
      id="config.logFile"
      envVar="LOG_FILE"
      bind:value={config.logFile}
      original={original.logFile} />
  </Col>
  <Col md={6}>
    <Input
      id="config.services.logFile"
      envVar="SERVICES_LOG_FILE"
      bind:value={config.services!.logFile}
      original={original.services?.logFile} />
  </Col>
  <Col md={6}>
    <Input
      id="config.httpLog"
      envVar="HTTP_LOG"
      bind:value={config.httpLog}
      original={original.httpLog} />
  </Col>
  <Col md={6}>
    <Input
      id="config.debugLog"
      envVar="DEBUG_LOG"
      bind:value={config.debugLog}
      original={original.debugLog} />
  </Col>
</Row>
<Row>
  <Col md={4}>
    <Input
      id="config.debug"
      envVar="DEBUG"
      type="select"
      bind:value={config.debug}
      original={original.debug} />
  </Col>
  <Col md={4}>
    <Input
      id="config.quiet"
      envVar="QUIET"
      type="select"
      bind:value={config.quiet}
      original={original.quiet} />
  </Col>
  <Col md={4}>
    <Input
      id="config.noUploads"
      envVar="NO_UPLOADS"
      type="select"
      bind:value={config.noUploads}
      original={original.noUploads} />
  </Col>
</Row>
<Row>
  <Col md={4}>
    <Input
      envVar="MAX_BODY"
      id="config.maxBody"
      type="number"
      bind:value={config.maxBody}
      min={500}
      max={500000}
      original={original.maxBody}>
      {#snippet post()}
        <InputGroupText>{$_('words.select-option.bytes')}</InputGroupText>
      {/snippet}
    </Input>
  </Col>
  <Col md={4}>
    <Input
      envVar="LOG_FILE_MB"
      id="config.logFileMb"
      type="number"
      min={1}
      max={999}
      bind:value={config.logFileMb}
      original={original.logFileMb}>
      {#snippet post()}
        <InputGroupText>{$_('words.select-option.megabytes')}</InputGroupText>
      {/snippet}
    </Input>
  </Col>
  <Col md={4}>
    <Input
      envVar="LOG_FILES"
      id="config.logFiles"
      type="number"
      min={0}
      bind:value={config.logFiles}
      original={original.logFiles} />
  </Col>
</Row>
