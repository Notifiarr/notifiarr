<script lang="ts" module>
  import { faHouseLock } from '@fortawesome/sharp-duotone-regular-svg-icons'
  export const page = {
    id: 'TrustProfile',
    i: faHouseLock,
    c1: 'steelblue',
    c2: 'darkblue',
    d1: 'lightsteelblue',
    d2: 'white',
  }
</script>

<script lang="ts">
  import { CardBody, Row, Col } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Input from '../../includes/Input.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { AuthType as Auth } from '../../api/notifiarrConfig'
  import Footer from '../../includes/Footer.svelte'
  import { onMount } from 'svelte'
  import Header from './Header.svelte'

  // Form state, this is what we're sending to the backend.
  const form = $state({
    username: $profile?.username || '',
    password: '',
    authType: $profile?.upstreamType || Auth.password,
    upstreams: $profile?.config.upstreams?.join(' ') || '',
    newPass: '',
    header: $profile?.config.uiPassword.startsWith('webauth:')
      ? $profile?.config.uiPassword.split(':')[1]
      : '',
  })

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

  const saveDisabled = $derived(
    // If current auth type is password/website, make sure their password is entered.
    ([Auth.password, Auth.website].includes($profile?.upstreamType) &&
      (!form.password || form.password.length < 5)) ||
      // If selected auth type is password and they entered a new password, make sure it's at least 9 characters.
      (form.authType === Auth.password && form.newPass && form.newPass.length < 9) ||
      // If they changed the auth type to password from something else, make sure they entered a new password.
      (form.authType === Auth.password &&
        $profile?.upstreamType !== Auth.password &&
        form.newPass.length < 9) ||
      // Make sure they picked a header if header auth type is selected.
      (form.authType === Auth.header && !form.header) ||
      // Make sure they didn't enter a username that's not allowed.
      form.username == 'webauth' ||
      form.username == 'noauth' ||
      form.username == 'website' ||
      form.username.includes(':') ||
      // Make sure the auth type changed, or upstreams changed.
      (form.authType === Auth.noauth &&
        $profile?.upstreamType === Auth.noauth &&
        form.upstreams === $profile?.upstreamIp) ||
      // Enable the save button.
      false,
  )
</script>

<Header {addit} />
<!-- We use a form here so you can press enter in a password field to save. -->
<form onsubmit={submit}>
  <CardBody class="pt-0 mt-0">
    <!-- Authorization Section -->
    <h4>{$_('profile.title.Authorization')}</h4>
    <Row>
      <Col md={6}>
        <Input
          id="profile.authType"
          type="select"
          bind:value={form.authType}
          options={[
            { value: Auth.password, name: $_('profile.authType.options.password') },
            { value: Auth.website, name: $_('profile.authType.options.website') },
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

    <!-- Authentication Section -->
    <h4>{$_('profile.title.Authentication')}</h4>
    {#if [Auth.password, Auth.website].includes(form.authType)}
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
    {:else if form.authType !== Auth.website}
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

    <!-- Current Password Section, shows up any time a password is configured in the backend. -->
    {#if [Auth.password, Auth.website].includes($profile?.upstreamType)}
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

<style>
  .fortune {
    white-space: pre-wrap;
  }
</style>
