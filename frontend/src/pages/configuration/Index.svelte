<script lang="ts" module>
  import { faFlaskGear } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'Configuration',
    i: faFlaskGear,
    c1: 'slategray',
    c2: 'darkgreen',
    d1: 'gainsboro',
    d2: 'lime',
  }
</script>

<script lang="ts">
  import { CardBody, InputGroupText, Row, Col } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Input from '../../includes/Input.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import { onMount } from 'svelte'
  import Header from '../../includes/Header.svelte'

  // Local state that syncs with profile store.
  const c = $derived({ ...$profile.config })
  // Convert array to newline-separated string for textarea
  let extraKeys = $derived(c.extraKeys?.join('\n') || '')
  const rows = $derived(
    extraKeys.split('\n').length > 10 ? 10 : extraKeys.split('\n').length,
  )

  // Handle form submission
  function submit() {
    c.extraKeys = extraKeys.split('\n').filter(key => key.trim() !== '')
    profile.writeConfig(c)
  }

  // Clear the status messages when the component unmounts.
  onMount(() => () => profile.clearStatus())
</script>

<Header {page} badge={$_('phrases.Version', { values: { version: c.version } })} />
<CardBody>
  <!-- General Section -->
  <h4 class="mb-2">{$_('config.titles.General')}</h4>
  <Input id="config.apiKey" type="password" bind:value={c.apiKey} />
  <Input id="config.extraKeys" type="textarea" bind:value={extraKeys} {rows} />
  <Input id="config.hostId" bind:value={c.hostId} />

  <!-- Network Section -->
  <h4 class="mb-2">{$_('config.titles.Network')}</h4>
  <Row>
    <Col md={6}>
      <Input id="config.bindAddr" bind:value={c.bindAddr} />
    </Col>
    <Col md={6}>
      <Input id="config.urlbase" bind:value={c.urlbase} />
    </Col>
  </Row>
  <Row>
    <Col md={6}>
      <Input
        id="config.timeout"
        type="select"
        bind:value={c.timeout}
        options={[
          { value: '15s', name: '15 ' + $_('words.select-option.seconds') },
          { value: '30s', name: '30 ' + $_('words.select-option.seconds') },
          { value: '1m0s', name: '1 ' + $_('words.select-option.minute') },
          { value: '2m0s', name: '2 ' + $_('words.select-option.minutes') },
          { value: '3m0s', name: '3 ' + $_('words.select-option.minutes') },
        ]} />
    </Col>
    <Col md={6}>
      <Input id="config.retries" type="number" bind:value={c.retries} min={0} />
    </Col>
  </Row>

  <!-- System Section -->
  <h4 class="mb-2">{$_('config.titles.System')}</h4>
  <Row>
    <Col md={4}>
      <Input id="config.serial" type="select" bind:value={c.serial} />
    </Col>
    {#if $profile.isWindows}
      <Col md={$profile.clientInfo?.user.devAllowed ? 4 : 8}>
        <Input
          id="config.autoUpdate"
          type="select"
          bind:value={c.autoUpdate}
          options={[
            { value: 'off', name: $_('words.select-option.Disabled') },
            { value: 'daily', name: $_('words.select-option.Daily') },
            { value: '12h', name: $_('phrases.Every12Hours') },
            { value: '6h', name: $_('phrases.Every6Hours') },
            { value: '3h', name: $_('phrases.Every3Hours') },
          ]} />
      </Col>
      {#if $profile.clientInfo?.user.devAllowed}
        <Col md={4}>
          <Input id="config.unstableCh" type="select" bind:value={c.unstableCh} />
        </Col>
      {/if}
    {:else}
      <Col md={4}>
        <Input
          type="select"
          id="config.fileMode"
          bind:value={c.fileMode}
          options={[
            { value: '0600', name: '0600 -rw-------' },
            { value: '0640', name: '0640 -rw-r-----' },
            { value: '0644', name: '0644 -rw-r--r--' },
            { value: '0604', name: '0604 -rw----r--' },
            { value: '0660', name: '0660 -rw-rw----' },
            { value: '0664', name: '0664 -rw-rw-r--' },
          ]} />
      </Col>
      <Col md={4}>
        <Input id="config.apt" type="select" bind:value={c.apt} />
      </Col>
    {/if}
  </Row>

  <!-- SSL Section -->
  <h4 class="mb-2">{$_('config.titles.SSLConfiguration')}</h4>
  <Row>
    <Col md={6}>
      <Input id="config.sslKeyFile" bind:value={c.sslKeyFile} />
    </Col>
    <Col md={6}>
      <Input id="config.sslCertFile" bind:value={c.sslCertFile} />
    </Col>
  </Row>

  <!-- Services Section -->
  <h4 class="mb-2">{$_('config.titles.Services')}</h4>
  <Row>
    <Col md={4}>
      <Input
        id="config.services.enabled"
        type="select"
        bind:value={c.services!.disabled} />
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
        bind:value={c.services!.parallel} />
    </Col>
    <Col md={4}>
      <Input
        id="config.services.interval"
        type="select"
        bind:value={c.services!.interval}
        options={[
          { value: '5m0s', name: '5 ' + $_('words.select-option.minutes') },
          { value: '10m0s', name: '10 ' + $_('words.select-option.minutes') },
          { value: '15m0s', name: '15 ' + $_('words.select-option.minutes') },
          { value: '20m0s', name: '20 ' + $_('words.select-option.minutes') },
          { value: '30m0s', name: '30 ' + $_('words.select-option.minutes') },
        ]} />
    </Col>
  </Row>

  <!-- Logging Section -->
  <h4 class="mb-2">{$_('config.titles.Logging')}</h4>
  <Row>
    <Col md={6}>
      <Input id="config.logFile" bind:value={c.logFile} />
    </Col>
    <Col md={6}>
      <Input id="config.services.logFile" bind:value={c.services!.logFile} />
    </Col>
    <Col md={6}>
      <Input id="config.httpLog" bind:value={c.httpLog} />
    </Col>
    <Col md={6}>
      <Input id="config.debugLog" bind:value={c.debugLog} />
    </Col>
  </Row>
  <Row>
    <Col md={4}>
      <Input id="config.debug" type="select" bind:value={c.debug} />
    </Col>
    <Col md={4}>
      <Input id="config.quiet" type="select" bind:value={c.quiet} />
    </Col>
    <Col md={4}>
      <Input id="config.noUploads" type="select" bind:value={c.noUploads} />
    </Col>
  </Row>
  <Row>
    <Col md={4}>
      <Input
        id="config.maxBody"
        type="number"
        bind:value={c.maxBody}
        min={500}
        max={500000}>
        {#snippet post()}
          <InputGroupText>
            {$_('words.select-option.bytes')}
          </InputGroupText>
        {/snippet}
      </Input>
    </Col>
    <Col md={4}>
      <Input
        id="config.logFileMb"
        type="number"
        min={1}
        max={999}
        bind:value={c.logFileMb}>
        {#snippet post()}
          <InputGroupText>
            {$_('words.select-option.megabytes')}
          </InputGroupText>
        {/snippet}
      </Input>
    </Col>
    <Col md={4}>
      <Input id="config.logFiles" type="number" min={0} bind:value={c.logFiles} />
    </Col>
  </Row>
</CardBody>
<Footer {submit} />
