<script lang="ts">
  import { Card, CardHeader, CardBody, Row, Col } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Input from '../../includes/Input.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import { AuthType as Auth } from '../../api/notifiarrConfig'
  import Footer from '../../includes/Footer.svelte'
  import { onMount } from 'svelte'
  import Fa from '../../includes/Fa.svelte'
  import { faSplotch, faUnlockAlt } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { slide } from 'svelte/transition'

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

  let showMore = false
  const toggleMore = (e: Event) => (e.preventDefault(), (showMore = !showMore))

  async function submit(e: Event) {
    e.preventDefault()
    await profile.trustProfile(form)
    form.newPass = form.password = ''
  }

  function addit(e: Event) {
    e.preventDefault()
    // Only add the upstream if it's not already in the list.
    if (!form.upstreams.split(/\s+/).includes($profile?.upstreamIp))
      form.upstreams += form.upstreams ? ' ' + $profile?.upstreamIp : $profile?.upstreamIp
  }

  // Clear the status messages when the component unmounts.
  onMount(() => () => profile.clearStatus())

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
    (profile.status != '' && !profile.success) ||
    // Enable the save button.
    false
</script>

<Card class="mb-2">
  <CardHeader>
    <h2><Fa icon={faUnlockAlt} /> {$_('navigation.titles.TrustProfile')}</h2>
    <p class="text-muted">{@html $_('profile.phrase.ProfileDescription')}</p>
  </CardHeader>

  <form onsubmit={submit}>
    <!-- Bullet Points-->
    <CardBody>
      <ul class="mb-2 list-unstyled">
        <li>
          <Fa icon={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
          {@html $_('profile.phrase.PasswordOnlyUsed')}
        </li>
        <li>
          <Fa icon={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
          {@html $_('profile.phrase.HeaderOnlyUsed')}
        </li>
        <li>
          <Fa icon={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
          {@html $_('profile.phrase.AuthTypeChange')}
        </li>
        <li>
          <Fa icon={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
          {@html $_('profile.phrase.SeeWikiForAuthProxyHelp')}
        </li>

        <!-- Missing/Add upstream section -->
        {#if $profile?.upstreamAllowed}
          <li>
            <Fa icon={faSplotch} c1="maroon" c2="salmon" d1="salmon" d2="maroon" />
            <b><T id="profile.phrase.ProxyAuthDisabled" {upstreamIp} /></b>
            <a href="#addit" onclick={addit}>
              {$_('profile.phrase.Addit')}
            </a>
          </li>
        {/if}

        <!-- Show more toggle -->
        <li>
          <Fa icon={faSplotch} c1="darkblue" d1="lightblue" c2="royalblue" />
          <a href="#toggle-more" onclick={toggleMore}>
            {showMore ? $_('profile.phrase.ShowMeLess') : $_('profile.phrase.ShowMeMore')}
          </a>
        </li>
      </ul>

      <!-- Show more section -->
      <Row>
        {#if showMore}
          <div transition:slide={{ duration: 900 }}>
            <Card body color="warning-subtle" class="my-1 pb-0">
              <p>{@html $_('profile.phrase.ShowMoreDesc')}</p>
              <h5>{$_('profile.authType.options.password')}</h5>
              <p>{@html $_('profile.authType.option_desc.password')}</p>
              <h5>{$_('profile.authType.options.header')}</h5>
              <p>{@html $_('profile.authType.option_desc.header')}</p>
              <h5>{$_('profile.authType.options.noauth')}</h5>
              <p>{@html $_('profile.authType.option_desc.noauth')}</p>
            </Card>
          </div>
        {/if}
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
    </CardBody>
    <!-- Form success/error section -->
    <Footer
      {submit}
      successText="profile.phrase.ProfileUpdated"
      saveButtonDescription="profile.phrase.EnterCurrentPassword"
      saveButtonText="buttons.SaveChanges"
      {saveDisabled}>
      <!-- Fortune section -->
      <hr />
      <pre class="p-3 rounded fortune">{$profile.fortune}</pre>
    </Footer>
  </form>
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
