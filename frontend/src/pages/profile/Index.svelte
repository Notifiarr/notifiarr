<script lang="ts" module>
  export const page = { id: 'TrustProfile' }
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
  import { nav } from '../../navigation/nav.svelte'

  // Form state, this is what we're sending to the backend.
  const form = $state({
    username: $profile?.username || '',
    password: '',
    authType: $profile?.upstreamType || Auth.password,
    upstreams: $profile?.config.upstreams?.join(' ') || '',
    newPass: '',
    header: $profile?.upstreamHeader || '',
  })

  // Native <select> may stringify enum values; keep comparisons numeric.
  const authType = $derived(Number(form.authType) as Auth)
  const origUpstreams = $derived($profile?.config.upstreams?.join(' ') || '')
  const reservedUser = $derived(
    form.username == 'webauth' ||
      form.username == 'noauth' ||
      form.username == 'website' ||
      form.username.includes(':'),
  )

  async function submit(e: Event) {
    e.preventDefault()
    const ok = await profile.trustProfile({ ...form, authType })
    if (ok) form.newPass = form.password = ''
  }

  function addit(e: Event) {
    e.preventDefault()
    // Only add the upstream if it's not already in the list.
    if (!form.upstreams.split(/\s+/).includes($profile?.upstreamIp))
      form.upstreams += form.upstreams ? ' ' + $profile?.upstreamIp : $profile?.upstreamIp
  }

  // Clear the status messages when the component unmounts.
  onMount(() => () => {
    profile.clearStatus()
    nav.formChanged = false
  })

  $effect(() => {
    nav.formChanged =
      authType !== $profile?.upstreamType ||
      form.upstreams !== origUpstreams ||
      form.header !== ($profile?.upstreamHeader || '') ||
      form.username !== ($profile?.username || '') ||
      !!form.newPass
  })

  const saveDisabled = $derived(
    // Current password/website auth requires the current password to save.
    ([Auth.password, Auth.website].includes($profile?.upstreamType) &&
      (!form.password || form.password.length < 5)) ||
      // New local password must be at least 9 characters when provided.
      (authType === Auth.password && form.newPass.length > 0 && form.newPass.length < 9) ||
      // Switching to password requires a new password and a usable username.
      (authType === Auth.password &&
        $profile?.upstreamType !== Auth.password &&
        form.newPass.length < 9) ||
      (authType === Auth.password && (!form.username || reservedUser)) ||
      // Header auth requires a header; noauth may omit it.
      (authType === Auth.header && !form.header) ||
      !nav.formChanged,
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
    {#if authType === Auth.header || authType === Auth.noauth}
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
    {:else if authType === Auth.password}
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
