<script lang="ts">
  import {
    Card,
    CardHeader,
    CardBody,
    CardFooter,
    FormGroup,
    Input,
    Label,
    InputGroup,
    InputGroupText,
    Button,
    Alert,
    Row,
    Col,
    Icon,
  } from '@sveltestrap/sveltestrap'
  import { profile } from '../api/profile'

  // Local state that syncs with profile store.
  $: c = { ...$profile.config }
  // Convert array to newline-separated string for textarea
  $: extraKeysText = c.extraKeys?.join('\n') + '\n' || ''

  // Helper function to toggle password visibility
  let showPassword = false
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
    c.extraKeys = extraKeysText.split('\n').filter(key => key.trim() !== '')

    try {
      await profile.updateConfig(c)
      submitSuccess = true
    } catch (error) {
      submitError =
        error instanceof Error ? error.message : 'An unknown error occurred'
    } finally {
      isSubmitting = false
    }
  }
</script>

<Card class="mb-4">
  <CardHeader><h2>Configuration</h2></CardHeader>
  <CardBody>
    <form on:submit={submit}>
      <!-- General Section -->
      <h3 class="mb-4">General</h3>

      <FormGroup>
        <Label for="apiKey">API Key</Label>
        <InputGroup>
          <Input
            type={showPassword ? 'text' : 'password'}
            id="apiKey"
            bind:value={c.apiKey}
            placeholder="Enter your Notifiarr.com API key" />
          <InputGroupText>
            <Button
              color="link"
              class="p-0"
              on:click={() => (showPassword = !showPassword)}>
              <Icon name={showPassword ? 'eye-slash' : 'eye'} />
            </Button>
          </InputGroupText>
        </InputGroup>
        <small class="text-muted">
          API key from your Notifiarr.com account. Find it at Notifiarr.com =>
          Profile => API Keys.
        </small>
      </FormGroup>

      <FormGroup>
        <Label for="extraKeys">Extra Keys</Label>
        <Input
          type="textarea"
          id="extraKeys"
          bind:value={extraKeysText}
          placeholder="Enter additional API keys (one per line)"
          rows={extraKeysText.split('\n').length > 10
            ? 10
            : extraKeysText.split('\n').length} />
        <small class="text-muted">
          Additional API keys for third-party integrations. Separate with
          newlines.
        </small>
      </FormGroup>

      <FormGroup>
        <Label for="bindAddr">Bind Address</Label>
        <Input
          type="text"
          id="bindAddr"
          bind:value={c.bindAddr}
          placeholder="0.0.0.0:5454" />
        <small class="text-muted">
          IP and port the app will listen on. Use 0.0.0.0 to listen on all
          interfaces.
        </small>
      </FormGroup>

      <!-- SSL Section -->
      <h3 class="mt-5 mb-4">SSL Configuration</h3>

      <FormGroup>
        <Label for="sslKeyFile">SSL Key File</Label>
        <Input
          id="sslKeyFile"
          bind:value={c.sslKeyFile}
          placeholder="Path to SSL key file" />
      </FormGroup>

      <FormGroup>
        <Label for="sslCertFile">SSL Certificate File</Label>
        <Input
          id="sslCertFile"
          bind:value={c.sslCertFile}
          placeholder="Path to SSL certificate file" />
      </FormGroup>

      <!-- Services Section -->
      <h3 class="mt-5 mb-4">Services</h3>

      <FormGroup>
        <Label for="servicesEnabled">Service Checks</Label>
        <Input
          type="select"
          id="servicesEnabled"
          bind:value={c.services!.disabled}>
          <option value={false}>Enabled</option>
          <option value={true}>Disabled</option>
        </Input>
      </FormGroup>

      <FormGroup>
        <Label for="servicesParallel">Parallel Checks</Label>
        <Input
          type="select"
          id="servicesParallel"
          bind:value={c.services!.parallel}>
          {#each Array(5) as _, i}
            <option value={i + 1}>{i + 1}</option>
          {/each}
        </Input>
      </FormGroup>

      <FormGroup>
        <Label for="servicesInterval">Update Interval</Label>
        <Input
          type="select"
          id="servicesInterval"
          bind:value={c.services!.interval}>
          <option value="5m">5 minutes</option>
          <option value="10m">10 minutes</option>
          <option value="15m">15 minutes</option>
          <option value="20m">20 minutes</option>
          <option value="30m">30 minutes</option>
        </Input>
      </FormGroup>

      <!-- Logging Section -->
      <h3 class="mt-5 mb-4">Logging</h3>

      <Row>
        <Col md={6}>
          <FormGroup>
            <Label for="debug">Debug Logging</Label>
            <Input type="select" id="debug" bind:value={c.debug}>
              <option value={false}>Disabled</option>
              <option value={true}>Enabled</option>
            </Input>
          </FormGroup>
        </Col>
        <Col md={6}>
          <FormGroup>
            <Label for="quiet">Quiet Logging</Label>
            <Input type="select" id="quiet" bind:value={c.quiet}>
              <option value={false}>Disabled</option>
              <option value={true}>Enabled</option>
            </Input>
          </FormGroup>
        </Col>
      </Row>

      <FormGroup>
        <Label for="logFile">Application Log File</Label>
        <Input
          id="logFile"
          bind:value={c.logFile}
          placeholder="/path/to/notifiarr.log" />
      </FormGroup>

      <FormGroup>
        <Label for="httpLog">HTTP Log File</Label>
        <Input
          id="httpLog"
          bind:value={c.httpLog}
          placeholder="/path/to/http-notifiarr.log" />
      </FormGroup>

      <FormGroup>
        <Label for="debugLog">Debug Log File</Label>
        <Input
          id="debugLog"
          bind:value={c.debugLog}
          placeholder="/path/to/debug-notifiarr.log" />
      </FormGroup>

      <Row>
        <Col md={4}>
          <FormGroup>
            <Label for="maxBody">Max Body Size</Label>
            <Input
              type="number"
              id="maxBody"
              bind:value={c.maxBody}
              min={500}
              max={500000} />
          </FormGroup>
        </Col>
        <Col md={4}>
          <FormGroup>
            <Label for="logFileMb">Log File Size (MB)</Label>
            <Input
              type="number"
              id="logFileMb"
              bind:value={c.logFileMb}
              min={1}
              max={999} />
          </FormGroup>
        </Col>
        <Col md={4}>
          <FormGroup>
            <Label for="logFiles">Log File Count</Label>
            <Input
              type="number"
              id="logFiles"
              bind:value={c.logFiles}
              min={0}
              max={999} />
          </FormGroup>
        </Col>
      </Row>

      <div class="mt-4">
        <Button color="primary" type="submit" disabled={isSubmitting}>
          {#if isSubmitting}
            Saving...
          {:else}
            Save Configuration
          {/if}
        </Button>
        {#if submitError}
          <Alert color="danger" class="mt-2" dismissible>{submitError}</Alert>
        {:else if submitSuccess}
          <Alert color="success" class="mt-2" dismissible
            >Configuration saved successfully.</Alert>
        {/if}
      </div>
    </form>
  </CardBody>
  <CardFooter>
    <small class="text-muted"
      >Configure your Notifiarr client settings here</small>
  </CardFooter>
</Card>

<style>
  :global(.form-group) {
    margin-bottom: 1rem;
  }

  h3 {
    color: #6c757d;
    font-size: 1.5rem;
    font-weight: 500;
  }

  :global(.input-group-text) {
    background-color: transparent;
    border: none;
    padding-right: 0;
  }

  :global(.btn-link) {
    color: #6c757d;
  }

  :global(.btn-link:hover) {
    color: #5a6268;
  }
</style>
