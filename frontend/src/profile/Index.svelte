<script lang="ts">
  import {
    Card,
    CardHeader,
    CardBody,
    CardFooter,
    Form,
    FormGroup,
    Label,
    Input,
    InputGroup,
    InputGroupText,
    Button,
    Alert,
    Table,
    Row,
    Col,
    Icon,
  } from '@sveltestrap/sveltestrap'
  import { profile } from '../api/profile'

  // Form state
  $: username = $profile?.username || ''
  $: upstreams = $profile?.config.upstreams || []
  let authType = 'password'
  let currentPassword = ''
  let newPassword = ''
  let authHeader = ''
  let showPassword = false
  let formSubmitted = false
  let formError = ''
  let formSuccess = false

  // Mock headers for demonstration (would come from API in real implementation)
  const headers = [
    { name: 'x-webauth-user', value: 'admin' },
    { name: 'remote-user', value: 'admin' },
    { name: 'x-auth-username', value: 'admin' },
  ]

  // Mock function to toggle password visibility
  function togglePasswordVisibility() {
    showPassword = !showPassword
  }

  // Mock function to save profile changes
  function saveProfileChanges() {
    formSubmitted = true

    // Form validation
    if (authType === 'password' && !currentPassword) {
      formError = 'You must enter your current password to make changes.'
      formSuccess = false
      return
    }

    if (authType === 'password' && newPassword && newPassword.length < 9) {
      formError = 'New password must be at least 9 characters.'
      formSuccess = false
      return
    }

    // Success case (would actually call API in real implementation)
    formError = ''
    formSuccess = true
  }
</script>

<Card class="mb-4">
  <CardHeader>
    <h2><Icon name="unlock-alt" /> Trust Profile</h2>
  </CardHeader>
  <CardBody>
    <p>This page controls how you log into this Notifiarr client application.</p>

    <ul class="mb-4">
      <li>
        <Icon name="star" class="text-secondary" />
        Username and Password are only used if Auth Type is set to Password.
      </li>
      <li>
        <Icon name="star" class="text-secondary" />
        Auth Header is only used if Auth Type is <em>not</em> set to Password.
      </li>
      <li>
        <Icon name="star" class="text-secondary" />
        If Auth Type changes, log out to complete the process.
      </li>
    </ul>

    {#if formSubmitted}
      {#if formError}
        <Alert toggle={() => (formSubmitted = false)} color="danger" dismissible>
          {formError}
        </Alert>
      {/if}
      {#if formSuccess}
        <Alert toggle={() => (formSubmitted = false)} color="success" dismissible>
          Profile updated successfully!
        </Alert>
      {/if}
    {/if}

    <Table bordered responsive>
      <thead>
        <tr>
          <th>Setting</th>
          <th>Current Value</th>
          <th>New Value</th>
        </tr>
      </thead>
      <tbody>
        <!-- Auth Type -->
        <tr>
          <td>
            <span class="fw-bold">Auth Type</span>
            <Icon
              name="question-circle"
              class="ms-2 text-info"
              title="Select how to authenticate with the client"
            />
          </td>
          <td
            >{authType === 'password'
              ? 'Password: Local Username'
              : authType === 'header'
                ? 'Auth Proxy: Header'
                : 'No Password'}</td
          >
          <td>
            <Input type="select" name="authType" id="authType" bind:value={authType}>
              <option value="password">Password: Local Username</option>
              <option value="header">Auth Proxy: Header</option>
              <option value="nopass">No Password (danger)</option>
            </Input>
          </td>
        </tr>

        <!-- Upstreams -->
        <tr>
          <td>
            <span class="fw-bold">Upstreams</span>
            <Icon
              name="question-circle"
              class="ms-2 text-info"
              title="IP addresses allowed to connect"
            />
          </td>
          <td>127.0.0.1 ::1</td>
          <td>
            <InputGroup>
              <Input
                type="text"
                name="upstreams"
                id="upstreams"
                placeholder="127.0.0.1 ::1"
                bind:value={upstreams}
              />
            </InputGroup>
          </td>
        </tr>

        <!-- Auth Header -->
        <tr>
          <td>
            <span class="fw-bold">Auth Header</span>
            <Icon
              name="question-circle"
              class="ms-2 text-info"
              title="Select the header containing your username"
            />
          </td>
          <td>none</td>
          <td>
            <Input
              type="select"
              name="authHeader"
              id="authHeader"
              bind:value={authHeader}
              disabled={authType !== 'header'}
            >
              {#each headers as header}
                <option value={header.name}>{header.name}: {header.value}</option>
              {/each}
            </Input>
          </td>
        </tr>

        <!-- Username -->
        <tr>
          <td>
            <span class="fw-bold">Username</span>
          </td>
          <td>{username || 'admin'}</td>
          <td>
            <InputGroup>
              <Input
                type="text"
                name="newUsername"
                id="newUsername"
                bind:value={username}
                disabled={authType !== 'password'}
              />
            </InputGroup>
          </td>
        </tr>

        <!-- Password -->
        <tr>
          <td>
            <span class="fw-bold">Password</span>
          </td>
          <td>
            {#if authType === 'password'}
              <InputGroup>
                <Input
                  type="password"
                  name="currentPassword"
                  id="currentPassword"
                  placeholder="enter current password"
                  bind:value={currentPassword}
                />
              </InputGroup>
            {/if}
            {#if authType !== 'password'}
              none
            {/if}
          </td>
          <td>
            <InputGroup>
              <Input
                type={showPassword ? 'text' : 'password'}
                name="newPassword"
                id="newPassword"
                placeholder="9 character minimum"
                bind:value={newPassword}
                disabled={authType !== 'password'}
              />
              <button
                class="btn btn-outline-secondary"
                type="button"
                aria-label="Toggle password visibility"
                on:click={() => (showPassword = !showPassword)}
              >
                <Icon name={showPassword ? 'eye-slash' : 'eye'} />
              </button>
            </InputGroup>
          </td>
        </tr>
      </tbody>
    </Table>

    <Button color="primary" on:click={saveProfileChanges}>Save Changes</Button>
    {#if authType === 'password'}
      <small class="ms-2 text-muted">You must enter your current password to make changes.</small>
    {/if}
  </CardBody>

  <CardFooter>
    <Row>
      <Col md="12">
        <h4>Fortune</h4>
        <pre class="bg-light p-3 rounded">
          "The best way to predict the future is to invent it." - Alan Kay
        </pre>
      </Col>
    </Row>
  </CardFooter>
</Card>
