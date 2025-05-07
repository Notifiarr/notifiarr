<script lang="ts">
  import {
    Card,
    CardHeader,
    CardBody,
    CardFooter,
    InputGroupText,
    Button,
    Alert,
    Row,
    Col,
    Icon,
    type InputType,
  } from '@sveltestrap/sveltestrap'
  import { profile } from '../api/profile'
  import ConfigInput from '../lib/Input.svelte'
  import { _ } from '../lib/Translate.svelte'

  // Local state that syncs with profile store.
  $: c = { ...$profile.config }
  // Convert array to newline-separated string for textarea
  $: extraKeys = c.extraKeys?.join('\n') + '\n' || ''
  $: rows = extraKeys.split('\n').length > 10 ? 10 : extraKeys.split('\n').length

  // Helper function to toggle password visibility
  let showPass = false
  let passType: InputType = 'password'
  function togglePass(e: Event | undefined) {
    e?.preventDefault()
    passType = (showPass = !showPass) ? 'text' : 'password'
  }

  // Form submission status
  let isSubmitting = false
  let submitError: string | null = null
  let submitSuccess = false

  // Handle form submission
  async function submit(event: Event) {
    event.preventDefault()
    isSubmitting = true
    submitError = null
    submitSuccess = false
    c.extraKeys = extraKeys.split('\n').filter(key => key.trim() !== '')

    try {
      await profile.updateConfig(c)
      submitSuccess = true
    } catch (error) {
      submitError =
        error instanceof Error
          ? error.message
          : $_('config.errors.AnUnknownErrorOccurred')
    } finally {
      isSubmitting = false
    }
  }
</script>

<div id="config" class="mb-2 pb-2">
  <Card>
    <CardHeader>
      <h2>{$_('config.titles.Configuration')}</h2>
      <p class="text-muted">{$_('phrases.ConfigureNotifiarrClientSettings')}</p>
    </CardHeader>
    <CardBody>
      <!-- General Section -->
      <h3 class="mb-2">{$_('config.titles.General')}</h3>
      <ConfigInput id="config.apiKey" type={passType} bind:value={c.apiKey}>
        <InputGroupText slot="post" class="toggle-button-container">
          <Button color="link" class="p-0 toggle-button" on:click={togglePass}>
            <Icon name={showPass ? 'eye-slash' : 'eye'} />
          </Button>
        </InputGroupText>
      </ConfigInput>
      <ConfigInput id="config.extraKeys" type="textarea" bind:value={extraKeys} {rows} />
      <ConfigInput id="config.hostId" bind:value={c.hostId} />

      <!-- Network Section -->
      <h3 class="mb-2">{$_('config.titles.Network')}</h3>
      <Row>
        <Col md={6}>
          <ConfigInput id="config.bindAddr" bind:value={c.bindAddr} />
        </Col>
        <Col md={6}>
          <ConfigInput id="config.urlbase" bind:value={c.urlbase} />
        </Col>
      </Row>

      <Row>
        <Col md={6}>
          <ConfigInput
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
          <ConfigInput id="config.retries" type="number" bind:value={c.retries} min={0} />
        </Col>
      </Row>

      <!-- System Section -->
      <h3 class="mb-2">{$_('config.titles.System')}</h3>
      <Row>
        <Col md={4}>
          <ConfigInput id="config.serial" type="select" bind:value={c.serial} />
        </Col>
        {#if $profile.isWindows}
          <Col md={$profile.clientInfo?.user.devAllowed ? 4 : 8}>
            <ConfigInput
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
              <ConfigInput
                id="config.unstableCh"
                type="select"
                bind:value={c.unstableCh} />
            </Col>
          {/if}
        {:else}
          <Col md={4}>
            <ConfigInput
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
            <ConfigInput id="config.apt" type="select" bind:value={c.apt} />
          </Col>
        {/if}
      </Row>

      <!-- SSL Section -->
      <h3 class="mb-2">{$_('config.titles.SSLConfiguration')}</h3>
      <Row>
        <Col md={6}>
          <ConfigInput id="config.sslKeyFile" bind:value={c.sslKeyFile} />
        </Col>
        <Col md={6}>
          <ConfigInput id="config.sslCertFile" bind:value={c.sslCertFile} />
        </Col>
      </Row>

      <!-- Services Section -->
      <h3 class="mb-2">{$_('config.titles.Services')}</h3>
      <Row>
        <Col md={4}>
          <ConfigInput
            id="config.services.enabled"
            type="select"
            bind:value={c.services!.disabled} />
        </Col>
        <Col md={4}>
          <ConfigInput
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
          <ConfigInput
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
      <h3 class="mb-2">{$_('config.titles.Logging')}</h3>
      <Row>
        <Col md={6}>
          <ConfigInput id="config.logFile" bind:value={c.logFile} />
        </Col>
        <Col md={6}>
          <ConfigInput id="config.services.logFile" bind:value={c.services!.logFile} />
        </Col>
        <Col md={6}>
          <ConfigInput id="config.httpLog" bind:value={c.httpLog} />
        </Col>
        <Col md={6}>
          <ConfigInput id="config.debugLog" bind:value={c.debugLog} />
        </Col>
      </Row>
      <Row>
        <Col md={4}>
          <ConfigInput id="config.debug" type="select" bind:value={c.debug} />
        </Col>
        <Col md={4}>
          <ConfigInput id="config.quiet" type="select" bind:value={c.quiet} />
        </Col>
        <Col md={4}>
          <ConfigInput id="config.noUploads" type="select" bind:value={c.noUploads} />
        </Col>
      </Row>
      <Row>
        <Col md={4}>
          <ConfigInput
            id="config.maxBody"
            type="number"
            bind:value={c.maxBody}
            min={500}
            max={500000}>
            <InputGroupText slot="post">
              {$_('words.select-option.bytes')}
            </InputGroupText>
          </ConfigInput>
        </Col>
        <Col md={4}>
          <ConfigInput
            id="config.logFileMb"
            type="number"
            min={1}
            max={999}
            bind:value={c.logFileMb}>
            <InputGroupText slot="post">
              {$_('words.select-option.megabytes')}
            </InputGroupText>
          </ConfigInput>
        </Col>
        <Col md={4}>
          <ConfigInput
            id="config.logFiles"
            type="number"
            min={0}
            bind:value={c.logFiles} />
        </Col>
      </Row>
    </CardBody>

    <CardFooter>
      <Row>
        <Col style="max-width: fit-content;">
          <Button
            size="lg"
            color="primary"
            type="submit"
            class="mt-1"
            disabled={isSubmitting}
            on:click={submit}>
            {#if isSubmitting}
              {$_('phrases.SavingConfiguration')}
            {:else}
              {$_('phrases.SaveConfiguration')}
            {/if}
          </Button>
        </Col>
        <Col>
          {#if submitError}
            <Alert color="danger" class="mt-1 mb-1" dismissible>{submitError}</Alert>
          {:else if submitSuccess}
            <Alert color="success" class="mt-1 mb-1" dismissible>
              {$_('phrases.ConfigurationSaved')}
            </Alert>
          {/if}
        </Col>
      </Row>
    </CardFooter>
  </Card>
</div>

<style>
  #config h3 {
    color: #031144;
    font-size: 1.5rem;
    font-weight: 500;
  }
</style>
