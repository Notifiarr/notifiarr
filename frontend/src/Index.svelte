<script lang="ts">
  import {
    Button,
    Card,
    CardBody,
    CardFooter,
    CardHeader,
    CardTitle,
    Col,
    Container,
    Icon,
    Input,
    Modal,
    ModalBody,
    ModalFooter,
    ModalHeader,
    Navbar,
    NavbarBrand,
    Row,
    Spinner,
  } from '@sveltestrap/sveltestrap'
  import logo from './assets/notifiarr.svg'
  import { profile, fetchProfile, updateProfile } from './api/profile'
  import { login } from './api/login'
  import Navigation from './Navigation.svelte'
  import { SvelteToast } from '@zerodevx/svelte-toast'
  import { isReady } from './lib/Translate.svelte'
  import { _ } from './lib/Translate.svelte'
  import { age } from './lib/util'
  import { checkReloaded, getUi } from './api/fetch'
  import { setLocale } from './lib/locale'
  import { onMount } from 'svelte'
  import { darkMode } from './lib/darkmode.svelte'
  import { urlbase } from './api/urlbase'

  let username = ''
  let password = ''
  let loginFailedMsg = ''
  let isLoading = false
  let navigate: Navigation
  let updateTimer: number
  let notification = ''
  let showReloadModal = false
  let showShutdownModal = false
  let showHelpModal = false
  let reload: any
  let shutdown: any

  async function timer() {
    if (!updateTimer) updateTimer = setInterval(timer, 1234)
    if (!$profile.updated || !$isReady) return
    const now = await Date.now()
    const diff = now - (await new Date($profile.updated).getTime())
    notification = $_('phrases.BackEndUpdated', { values: { age: age(diff) } })
  }

  $: if ($isReady && $profile?.loggedIn) timer()

  onMount(() => {
    const query = new URLSearchParams(window.location.search)
    if (query.get('lang')) setLocale(query.get('lang')!)
  })

  async function updateBackend() {
    clearInterval(updateTimer)
    updateTimer = 0
    notification = `<span class="text-warning">${$_('phrases.UpdatingBackEnd')}</span>`
    try {
      await updateProfile()
      timer()
    } catch (err) {
      notification = `<span class="text-danger">${$_('phrases.FailedToUpdateBackEnd', {
        values: { error: `${err}` },
      })}</span>`
    }
  }

  async function handleLogin(e: Event) {
    e.preventDefault()

    if (!username || !password) {
      loginFailedMsg = $_('config.errors.PleaseEnterBothUsernameAndPassword')
      return
    }

    isLoading = true
    loginFailedMsg = ''

    loginFailedMsg = (await login(username, password)) ?? ''
    if (!loginFailedMsg) {
      notification = $_('phrases.LoggedIn')
      timer()
    } else {
      loginFailedMsg = $_('config.errors.LoginFailed', {
        values: { error: loginFailedMsg },
      })
    }

    isLoading = false
  }

  const confirmReload = () => (showReloadModal = true)
  const confirmShutdown = () => (showShutdownModal = true)

  $: theme = $darkMode ? 'dark' : 'light'
  $: $darkMode
    ? window.document.body.classList.add('dark-mode')
    : window.document.body.classList.remove('dark-mode')
</script>

<svelte:head>
  <title>{$profile.loggedIn ? '' : `Login - `}Notifiarr Client</title>
  <link rel="icon" type="image/svg+xml" href={logo} />
</svelte:head>

<SvelteToast />

<main>
  <Container fluid class="mb-2">
    <Navbar {theme} class="mb-0 pb-0">
      {#if $profile.loggedIn}
        <span style="position: absolute; right: 0;" class="fs-3">
          <Icon
            name="bootstrap-reboot"
            class="text-success me-1"
            onclick={confirmReload} />
          <Icon name="power" class="text-danger me-2" onclick={confirmShutdown} />
        </span>
      {/if}
      <NavbarBrand href={$urlbase} onclick={e => navigate.goto(e, '')} class="mb-0 pb-0">
        <h1 class="m-0 lh-1" style="font-size: 40px;">
          <img src={logo} height="45" alt="Logo" />
          <span class="title-notifiarr">Notifiarr Client</span>
        </h1>
      </NavbarBrand>
    </Navbar>

    <!-- Notification Center-->
    <Row class="mt-0 mb-1 lh-1">
      <Col class="fs-6 fs-lighter ms-3 fst-italic">
        <Card color="transparent border-0" {theme}>
          <span class="text-nowrap">
            {#if $profile?.loggedIn}
              <Icon name="arrow-counterclockwise" onclick={updateBackend} />
            {/if}
            {@html notification}
          </span>
        </Card>
      </Col>
    </Row>

    <!-- Shutdown Confirmation Modal -->
    <Modal isOpen={showShutdownModal} {theme}>
      <ModalHeader>{$_('phrases.ConfirmShutdown')}</ModalHeader>
      {#if shutdown}
        <ModalBody>
          {#await shutdown() then result}
            {#if result.ok}
              <span class="text-danger">{$_('phrases.ShutdownSuccess')}</span>
            {:else}
              {$_('phrases.FailedToShutdown', { values: { error: result.body } })}
            {/if}
          {/await}
        </ModalBody>
      {:else}
        <ModalBody>{$_('phrases.ConfirmShutdownBody')}</ModalBody>
        <ModalFooter>
          <Button
            color="danger"
            onclick={() => (shutdown = async () => await getUi('shutdown', false))}>
            {$_('buttons.Confirm')}
          </Button>
          <Button color="secondary" onclick={() => (showShutdownModal = false)}>
            {$_('buttons.Cancel')}
          </Button>
        </ModalFooter>
      {/if}
    </Modal>

    <!-- Reload Confirmation Modal -->
    <Modal isOpen={showReloadModal} toggle={() => (showReloadModal = false)} {theme}>
      <ModalHeader>{$_('phrases.ConfirmReload')}</ModalHeader>
      {#if reload}
        {#await reload() then result}
          {#if result.ok}
            {#await checkReloaded()}
              <ModalBody><Spinner size="sm" /> {$_('phrases.Reloading')}</ModalBody>
            {:then}
              {updateBackend()}
              {(showReloadModal = false)}
              {(reload = null)}
            {:catch error}
              {(showReloadModal = false)}
              {(reload = null)}
              {(notification = `<span class="text-danger">${$_('phrases.FailedToReload', {
                values: { error: error.message },
              })}</span>`)}
            {/await}
          {:else}
            {(showReloadModal = false)}
            {clearInterval(updateTimer)}
            {(updateTimer = 0)}
            {(notification = `<span class="text-danger">${$_('phrases.FailedToReload', {
              values: { error: result.body },
            })}</span>`)}
            {(reload = null)}
          {/if}
        {/await}
      {:else}
        <ModalBody>{$_('phrases.ConfirmReloadBody')}</ModalBody>
        <ModalFooter>
          <Button
            color="danger"
            onclick={async () => (reload = async () => await getUi('reload', false))}>
            {$_('buttons.Confirm')}</Button>
          <Button color="secondary" onclick={() => (showReloadModal = false)}>
            {$_('buttons.Cancel')}
          </Button>
        </ModalFooter>
      {/if}
    </Modal>

    <!-- Login Help Modal -->
    <Modal isOpen={showHelpModal} toggle={() => (showHelpModal = false)} {theme}>
      <ModalHeader>{$_('phrases.LoginHelp')}</ModalHeader>
      <ModalBody>{@html $_('phrases.LoginHelpBody')}</ModalBody>
      <ModalFooter>
        <Button color="secondary" onclick={() => (showHelpModal = false)}>
          {$_('buttons.Close')}
        </Button>
      </ModalFooter>
    </Modal>

    <Row>
      <!-- Wait for translations to load. -->
      {#if !$isReady}
        <Col xs={{ size: 8, offset: 2 }} md={{ size: 4, offset: 4 }}>
          <Card outline {theme} color="notifiarr">
            <CardBody class="text-nowrap fs-3">
              <!-- This is before translations are loaded. This 'typo' is on purpose, sue me.-->
              <Spinner /> Translateratating!...</CardBody>
          </Card>
        </Col>
      {:else}
        {#await fetchProfile()}
          <!-- Wait for profile to load. -->
          <Col xs={{ size: 8, offset: 2 }} md={{ size: 4, offset: 4 }}>
            <Card outline {theme} color="notifiarr">
              <CardBody class="text-nowrap fs-3">
                <Spinner />{$_('phrases.Loading')}</CardBody>
            </Card>
          </Col>
        {:then}
          {#if $profile.loggedIn}
            <!-- This is the main page, after logging in. -->
            <Navigation bind:this={navigate} />
          {:else}
            <!-- This is the login page, before logging in. -->
            <Col xs={{ size: 8, offset: 2 }} md={{ size: 4, offset: 4 }}>
              <Card outline {theme} class="mt-2" color="notifiarr">
                <CardHeader>
                  <CardTitle>{$_('buttons.Login')}</CardTitle>
                </CardHeader>
                <CardBody>
                  <form onsubmit={handleLogin}>
                    <Input
                      type="text"
                      name="username"
                      id="username"
                      placeholder="Username"
                      bind:value={username} />
                    <Input
                      type="password"
                      name="password"
                      id="password"
                      placeholder="Password"
                      class="my-1"
                      bind:value={password} />
                    <Button
                      type="submit"
                      size="sm"
                      disabled={isLoading}
                      class="w-100"
                      color="success">
                      {#if isLoading}
                        <Spinner size="sm" />
                        <span class="fs-5">{$_('phrases.LoggingIn')}</span>
                      {:else}
                        <span class="fs-5">{$_('buttons.Login')}</span>
                      {/if}
                    </Button>
                  </form>
                  <CardFooter class="mt-2">
                    <a
                      href="#showhelp"
                      onclick={e => (e.preventDefault(), (showHelpModal = true))}>
                      {$_('phrases.LoginHelp')}
                    </a>
                    {#if loginFailedMsg}
                      • <span class="text-danger">{loginFailedMsg}</span>
                    {/if}
                  </CardFooter>
                </CardBody>
              </Card>
            </Col>
          {/if}
        {:catch error}
          <Col xs={{ size: 10, offset: 1 }} md={{ size: 6, offset: 3 }}>
            <!-- error fetching profile (ie. timeout) -->
            <Card outline body {theme} color="danger">
              <CardHeader><CardTitle>{$_('phrases.ERROR')}</CardTitle></CardHeader>
              <CardBody>{error.message}</CardBody>
              <CardFooter>{$_('phrases.TryRefreshingThePage')}</CardFooter>
            </Card>
          </Col>
        {/await}
      {/if}
    </Row>
  </Container>
</main>
