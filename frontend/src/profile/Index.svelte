<script lang="ts">
  import {
    Card,
    CardHeader,
    CardBody,
    CardFooter,
    Button,
    Alert,
    Row,
    Col,
    Icon,
    Fade,
    Spinner,
  } from '@sveltestrap/sveltestrap'
  import { profile, updateProfile } from '../api/profile'
  import Input from '../lib/Input.svelte'
  import T, { _ } from '../lib/Translate.svelte'
  import { AuthType as Auth } from '../api/notifiarrConfig'
  import { checkReloaded, postUi } from '../api/fetch'
  import { darkMode } from '../lib/darkmode.svelte'

  $: theme = $darkMode ? 'dark' : 'light'
  $: upstreamIp = '<span class="text-danger">' + $profile?.upstreamIp + '</span>' // goes into a translation.
  // Form state, this is what we're sending to the backend.
  $: form = {
    username: $profile?.username || '',
    password: '',
    authType: $profile?.upstreamType || Auth.password,
    upstreams: $profile?.config.upstreams?.join(' ') || '',
    newPass: '',
    header: $profile?.config.uiPassword.startsWith('webauth:')
      ? $profile?.config.uiPassword.split(':')[1]
      : '',
  }

  let formSubmitted = ''
  let formError: unknown
  let formSuccess = false
  let showMore = false

  // Mock function to save profile changes
  async function saveProfileChanges() {
    formSubmitted = $_('phrases.SavingConfiguration')
    const { ok, body } = await postUi('profile', JSON.stringify(form), false)
    if (!ok) {
      formError = body
      form.password = ''
      return
    }

    try {
      formSubmitted = $_('phrases.Reloading')
      await checkReloaded()
      await updateProfile()
    } catch (e) {
      formError = e
      form.newPass = form.password = ''
      return
    }

    formSuccess = true
    formSubmitted = $_('phrases.ConfigurationSaved')
    form.newPass = form.password = ''
  }

  function addit() {
    // Only add the upstream if it's not already in the list.
    if (!form.upstreams.split(/\s+/).includes($profile?.upstreamIp))
      form.upstreams += form.upstreams ? ' ' + $profile?.upstreamIp : $profile?.upstreamIp
  }

  $: saveDisabled =
    // If current auth type is password, make sure their password is entered.
    ($profile?.upstreamType === Auth.password &&
      (!form.password || form.password.length < 9)) ||
    // If selected auth type is password and they entered a new password, make sure it's at least 9 characters.
    (form.authType === Auth.password && form.newPass && form.newPass.length < 9) ||
    // Make sure they picked a header if header auth type is selected.
    (form.authType === Auth.header && !form.header) ||
    // Make sure they didn't enter a username that's not allowed.
    form.username == 'webauth' ||
    form.username == 'noauth' ||
    form.username.includes(':') ||
    // Cannot click save while the form is submitting.
    // This also locks it if the form had a submission error.
    (formSubmitted != '' && !formSuccess) ||
    // Enable the save button.
    false
</script>

<Card class="mb-4" {theme}>
  <CardHeader>
    <h2><Icon name="unlock-alt" /> {$_('navigation.titles.TrustProfile')}</h2>
    <p class="text-muted">{@html $_('profile.phrase.ProfileDescription')}</p>
  </CardHeader>

  <!-- Bullet Points-->
  <CardBody>
    <ul class="mb-2 list-unstyled">
      <li>
        <Icon name="star" class="text-secondary" />
        {@html $_('profile.phrase.PasswordOnlyUsed')}
      </li>
      <li>
        <Icon name="star" class="text-secondary" />
        {@html $_('profile.phrase.HeaderOnlyUsed')}
      </li>
      <li>
        <Icon name="star" class="text-secondary" />
        {@html $_('profile.phrase.AuthTypeChange')}
      </li>
      <li>
        <Icon name="star" class="text-secondary" />
        {@html $_('profile.phrase.SeeWikiForAuthProxyHelp')}
      </li>

      <!-- Missing/Add upstream section -->
      {#if !$profile?.upstreamAllowed}
        <li>
          <Icon name="star" class="text-danger" />
          <b><T id="profile.phrase.ProxyAuthDisabled" {upstreamIp} /></b>
          <a href="#addit" on:click|preventDefault={addit}>
            {$_('profile.phrase.Addit')}
          </a>
        </li>
      {/if}

      <!-- Show more toggle -->
      <li>
        <Icon name="star" class="text-primary" />
        <a href="#toggle-more" on:click|preventDefault={() => (showMore = !showMore)}>
          {showMore ? $_('profile.phrase.ShowMeLess') : $_('profile.phrase.ShowMeMore')}
        </a>
      </li>
    </ul>

    <!-- Show more section -->
    <Row>
      <Fade isOpen={showMore}>
        <Card body color="warning-subtle" class="my-1 pb-0">
          <p>{@html $_('profile.phrase.ShowMoreDesc')}</p>
          <h5>{$_('profile.authType.options.password')}</h5>
          <p>{@html $_('profile.authType.option_desc.password')}</p>
          <h5>{$_('profile.authType.options.header')}</h5>
          <p>{@html $_('profile.authType.option_desc.header')}</p>
          <h5>{$_('profile.authType.options.noauth')}</h5>
          <p>{@html $_('profile.authType.option_desc.noauth')}</p>
        </Card>
      </Fade>
    </Row>

    <!-- Authorization Section -->
    <h3 class="my-2">{$_('profile.title.Authorization')}</h3>
    <Row>
      <Col md={6}>
        <Input
          id="profile.authType"
          type="select"
          bind:value={form.authType}
          options={[
            { value: Auth.password, name: $_('profile.authType.options.password') },
            {
              value: Auth.header,
              name: $_('profile.authType.options.header'),
              disabled: !$profile?.upstreamAllowed,
            },
            {
              value: Auth.noauth,
              name: $_('profile.authType.options.noauth'),
              disabled: !$profile?.upstreamAllowed,
            },
          ]} />
      </Col>
      <Col md={6}>
        <Input
          id="profile.upstreams"
          type="text"
          bind:value={form.upstreams}
          placeholder={$_('profile.upstreams.placeholder')} />
      </Col>
    </Row>

    <!-- We use a form here so you can press enter in a password field to save. -->
    <form on:submit|preventDefault={saveProfileChanges}>
      <!-- Authentication Section -->
      {#if form.authType !== Auth.noauth}
        <h3 class="mb-2">{$_('profile.title.Authentication')}</h3>
        {#if form.authType === Auth.header}
          <Row>
            <Col md={8}>
              <Input id="profile.header" type="select" bind:value={form.header}>
                {#each Object.entries($profile?.headers || {}) as [key, value]}
                  {#each value! as val}
                    <option
                      value={key}
                      selected={form.header.toLowerCase() === key.toLowerCase()}>
                      {key} ({val})
                    </option>
                  {/each}
                {/each}
                {#if form.header === ''}
                  <option value={form.header} selected>(none)</option>
                {:else if !$profile?.headers?.[form.header]}
                  <option value={form.header} selected>
                    {form.header} (other)
                  </option>
                {/if}
              </Input>
            </Col>
          </Row>
        {:else if form.authType === Auth.password}
          <Row>
            <Col md={8}>
              <Input
                id="profile.newPass"
                name="noautofill"
                type="password"
                bind:value={form.newPass} />
            </Col>
          </Row>
          <Row>
            <Col md={8}>
              <Input id="profile.username" type="text" bind:value={form.username} />
            </Col>
          </Row>
        {/if}
      {/if}

      <!-- Current Password Section, shows up any time a password is configured in the backend. -->
      {#if $profile?.upstreamType === Auth.password}
        <Row>
          <Col md={8}>
            <Input
              id="profile.password"
              name="password"
              type="password"
              bind:value={form.password} />
          </Col>
        </Row>
      {/if}

      <Row>
        <!-- Save Changes Section -->
        <Col style="max-width: fit-content;">
          <Button color="primary" type="submit" disabled={saveDisabled}>
            {#if formSubmitted}
              {$_('phrases.SavingConfiguration')}
            {:else}
              {$_('buttons.SaveChanges')}
            {/if}
          </Button><br />
          {#if $profile?.upstreamType === Auth.password}
            <small class="ms-2 text-muted">
              {$_('profile.phrase.EnterCurrentPassword')}
            </small>
          {/if}
        </Col>

        <!-- Form success/error section -->
        <Col>
          {#if formSubmitted}
            {#if formError}
              <Alert
                toggle={() => (formSubmitted = '')}
                color="danger"
                dismissible
                class="submit-alert"
                closeClassName="submit-alert-close">
                {formError}
              </Alert>
            {:else if formSuccess}
              <Alert
                toggle={() => (formSubmitted = '')}
                color="success"
                dismissible
                class="submit-alert"
                closeClassName="submit-alert-close">
                {$_('profile.phrase.ProfileUpdated')}
              </Alert>
            {:else}
              <Alert color="warning" class="submit-alert">
                <Spinner size="sm" color="warning" />
                {formSubmitted}
              </Alert>
            {/if}
          {/if}
        </Col>
      </Row>
    </form>
  </CardBody>

  <!-- Fortune section -->
  <CardFooter>
    <h4>Fortune</h4>
    <pre class="p-3 rounded fortune">
      {$profile?.fortune || $_('profile.phrase.FortuneDefault')}
    </pre>
  </CardFooter>
</Card>

<style>
  .fortune {
    white-space: pre-wrap;
  }

  h3 {
    font-size: 1.5rem;
    font-weight: 500;
  }
</style>
