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
  import { profile } from './api/profile.svelte'
  import Navigation from './Navigation.svelte'
  import { SvelteToast } from '@zerodevx/svelte-toast'
  import T, { isReady, _ } from './includes/Translate.svelte'
  import { age, delay } from './includes/util'
  import { checkReloaded, getUi } from './api/fetch'
  import { setLocale } from './includes/locale/index.svelte'
  import { onMount } from 'svelte'
  import { theme as thm } from './includes/theme.svelte'
  import { urlbase } from './api/fetch'
  import Fa from 'svelte-fa'
  import {
    faArrowsRepeat,
    faRotate,
    faPowerOff,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'

  let username = ''
  let password = ''
  let loginFailedMsg = ''
  let isLoading = false
  let navigate: Navigation
  let notification = ''
  let showReloadModal = false
  let showShutdownModal = false
  let showHelpModal = false
  let reload: any
  let shutdown: any

  onMount(() => {
    const query = new URLSearchParams(window.location.search)
    if (query.get('lang')) setLocale(query.get('lang')!)
  })

  async function updateBackend(e?: Event) {
    e?.preventDefault()
    notification = `<span class="text-warning">${$_('phrases.UpdatingBackEnd')}</span>`
    try {
      await profile.refresh()
      await delay(3456)
      notification = ''
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

    loginFailedMsg = (await profile.login(username, password)) ?? ''
    if (!loginFailedMsg) {
      notification = $_('phrases.LoggedIn')
      await delay(4567)
      notification = ''
    } else {
      loginFailedMsg = $_('config.errors.LoginFailed', {
        values: { error: loginFailedMsg },
      })
    }

    isLoading = false
  }

  const confirmReload = (e: Event) => (e.preventDefault(), (showReloadModal = true))
  const confirmShutdown = (e: Event) => (e.preventDefault(), (showShutdownModal = true))
  $: theme = $thm
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
          <a href="#reload" onclick={confirmReload}>
            <Fa
              icon={faRotate}
              primaryColor="#33A000"
              secondaryColor="#33A5A4"
              class="me-1" />
          </a>
          <a href="#shutdown" onclick={confirmShutdown}>
            <Fa
              icon={faPowerOff}
              primaryColor="#AA4B65"
              secondaryColor="red"
              class="me-2" />
          </a>
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
              <a href="#reload" onclick={updateBackend}>
                <Fa
                  icon={faArrowsRepeat}
                  secondaryColor="#3cd2a5"
                  primaryColor="green"
                  class="me-1" />
              </a>
              {#if notification}
                {@html notification}
              {:else}
                <T id="phrases.BackEndUpdated" age={age(profile.updatedAge)} />
              {/if}
            {/if}
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
            <!-- reload success! -->
            {#await checkReloaded()}
              <!-- wait for reload to complete. -->
              <ModalBody><Spinner size="sm" /> {$_('phrases.Reloading')}</ModalBody>
            {:then}
              <!-- reload complete! -->
              {updateBackend()}
              {(showReloadModal = false)}
              {(reload = null)}
            {:catch error}
              <!-- error waiting for reload to complete. -->
              {(showReloadModal = false)}
              {(reload = null)}
              {(notification = `<span class="text-danger">${$_('phrases.FailedToReload', {
                values: { error: error.message },
              })}</span>`)}
            {/await}
          {:else}
            <!-- reload command failed. prob logged out. -->
            {(showReloadModal = false)}
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
        {#await profile.fetch()}
          <!-- Wait for profile to load. -->
          <Col xs={{ size: 8, offset: 2 }} md={{ size: 4, offset: 4 }}>
            <Card outline {theme} color="notifiarr">
              <CardBody class="text-nowrap fs-3">
                <Spinner /> {$_('phrases.Loading')}</CardBody>
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
