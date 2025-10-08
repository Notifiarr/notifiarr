<script lang="ts">
  import { Row, Col, InputGroupText } from '@sveltestrap/sveltestrap'
  import Input from '../../includes/Input.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import type { Config } from '../../api/notifiarrConfig'
  import BrowserModal from '../../includes/fileBrowser/BModal.svelte'

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
      bind:value={config.logFile}
      original={original.logFile}
      onclick={() => (logFileModal = true)} />
  </Col>
  <Col md={6}>
    <Input
      id="config.services.logFile"
      bind:value={config.services!.logFile}
      original={original.services?.logFile}
      onclick={() => (servicesLogModal = true)} />
  </Col>
  <Col md={6}>
    <Input
      id="config.httpLog"
      bind:value={config.httpLog}
      original={original.httpLog}
      onclick={() => (httpLogModal = true)} />
  </Col>
  <Col md={6}>
    <Input
      id="config.debugLog"
      bind:value={config.debugLog}
      original={original.debugLog}
      onclick={() => (debugLogModal = true)} />
  </Col>
</Row>
<Row>
  <Col md={4}>
    <Input
      id="config.debug"
      type="select"
      bind:value={config.debug}
      original={original.debug} />
  </Col>
  <Col md={4}>
    <Input
      id="config.quiet"
      type="select"
      bind:value={config.quiet}
      original={original.quiet} />
  </Col>
  <Col md={4}>
    <Input
      id="config.noUploads"
      type="select"
      bind:value={config.noUploads}
      original={original.noUploads} />
  </Col>
</Row>
<Row>
  <Col md={4}>
    <Input
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
      id="config.logFiles"
      type="number"
      min={0}
      bind:value={config.logFiles}
      original={original.logFiles} />
  </Col>
</Row>

<BrowserModal
  bind:isOpen={logFileModal}
  bind:value={config.logFile}
  title="config.logFile.label">
  <p><T id="config.logFile.description" /></p>
  <T id="config.logFile.tooltip" />
</BrowserModal>

<BrowserModal
  bind:isOpen={debugLogModal}
  bind:value={config.debugLog}
  title="config.debugLog.label">
  <p><T id="config.debugLog.description" /></p>
  <T id="config.debugLog.tooltip" />
</BrowserModal>

<BrowserModal
  bind:isOpen={httpLogModal}
  bind:value={config.httpLog}
  title="config.httpLog.label">
  <p><T id="config.httpLog.description" /></p>
  <T id="config.httpLog.tooltip" />
</BrowserModal>

<BrowserModal
  bind:isOpen={servicesLogModal}
  bind:value={config.services!.logFile}
  title="config.services.logFile.label">
  <p><T id="config.services.logFile.description" /></p>
  <T id="config.services.logFile.tooltip" />
</BrowserModal>
