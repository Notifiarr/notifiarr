<script lang="ts">
  import { Row, Col, InputGroupText } from '@sveltestrap/sveltestrap'
  import Input from '../../includes/Input.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import type { Config } from '../../api/notifiarrConfig'
  import BrowserModal from '../../includes/fileBrowser/BModal.svelte'
  import BButton from '../../includes/fileBrowser/BButton.svelte'

  type Props = { config: Config; original: Config }
  const { config = $bindable(), original }: Props = $props()

  let logFileModal = $state(false)
  let debugLogModal = $state(false)
  let httpLogModal = $state(false)
  let servicesLogModal = $state(false)
</script>

<!-- Logging Section -->
<h4>{$_('config.titles.Logging')}</h4>
<Row>
  <Col md={6}>
    <Input
      id="config.logFile"
      envVar="LOG_FILE"
      bind:value={config.logFile}
      original={original.logFile}>
      {#snippet post()}<BButton bind:isOpen={logFileModal} />{/snippet}
    </Input>
  </Col>
  <Col md={6}>
    <Input
      id="config.services.logFile"
      envVar="SERVICES_LOG_FILE"
      bind:value={config.services.logFile}
      original={original.services?.logFile}>
      {#snippet post()}<BButton bind:isOpen={servicesLogModal} />{/snippet}
    </Input>
  </Col>
  <Col md={6}>
    <Input
      id="config.httpLog"
      envVar="HTTP_LOG"
      bind:value={config.httpLog}
      original={original.httpLog}>
      {#snippet post()}<BButton bind:isOpen={httpLogModal} />{/snippet}
    </Input>
  </Col>
  <Col md={6}>
    <Input
      id="config.debugLog"
      envVar="DEBUG_LOG"
      bind:value={config.debugLog}
      original={original.debugLog}>
      {#snippet post()}<BButton bind:isOpen={debugLogModal} />{/snippet}
    </Input>
  </Col>
</Row>
<Row>
  <Col md={6}>
    <Input
      id="config.debug"
      envVar="DEBUG"
      type="select"
      bind:value={config.debug}
      original={original.debug} />
  </Col>
  <Col md={6}>
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
</Row>
<Row>
  <Col md={6}>
    <Input
      id="config.quiet"
      envVar="QUIET"
      type="select"
      bind:value={config.quiet}
      original={original.quiet} />
  </Col>
  <Col md={6}>
    <Input
      id="config.trace"
      envVar="TRACE"
      type="select"
      bind:value={config.trace}
      original={original.trace} />
  </Col>
</Row>
<Row>
  <Col md={4}>
    <Input
      id="config.noUploads"
      envVar="NO_UPLOADS"
      type="select"
      bind:value={config.noUploads}
      original={original.noUploads} />
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
</Row>

<BrowserModal
  bind:isOpen={logFileModal}
  bind:value={config.logFile}
  title="config.logFile.label"
  description="config.logFile.description">
  <T id="config.logFile.tooltip" />
</BrowserModal>

<BrowserModal
  bind:isOpen={debugLogModal}
  bind:value={config.debugLog}
  title="config.debugLog.label"
  description="config.debugLog.description">
  <T id="config.debugLog.tooltip" />
</BrowserModal>

<BrowserModal
  bind:isOpen={httpLogModal}
  bind:value={config.httpLog}
  title="config.httpLog.label"
  description="config.httpLog.description">
  <T id="config.httpLog.tooltip" />
</BrowserModal>

<BrowserModal
  bind:isOpen={servicesLogModal}
  bind:value={config.services.logFile}
  title="config.services.logFile.label"
  description="config.services.logFile.description">
  <T id="config.services.logFile.tooltip" />
</BrowserModal>
