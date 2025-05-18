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
    Row,
    Spinner,
  } from '@sveltestrap/sveltestrap'
  import logo from './assets/notifiarr.svg?inline'
  import { profile } from './api/profile.svelte'
  import Navigation from './navigation/Index.svelte'
  import { SvelteToast } from '@zerodevx/svelte-toast'
  import { isReady, _ } from './includes/Translate.svelte'
  import { delay } from './includes/util'
  import { setLocale } from './includes/locale/index.svelte'
  import { onMount } from 'svelte'
  import { theme } from './includes/theme.svelte'
  import MainHeader, { showMsg } from './header/Index.svelte'

  let username = $state('')
  let password = $state('')
  let loginFailedMsg = $state('')
  let isLoading = $state(false)
  let showHelpModal = $state(false)

  onMount(() => {
    const query = new URLSearchParams(window.location.search)
    if (query.get('lang')) setLocale(query.get('lang')!)
  })

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
      showMsg($_('phrases.LoggedIn'))
      await delay(4567)
      showMsg('')
    } else {
      loginFailedMsg = $_('config.errors.LoginFailed', {
        values: { error: loginFailedMsg },
      })
    }

    isLoading = false
  }
</script>

<svelte:head>
  <title>{$profile.loggedIn ? '' : `Login - `}Notifiarr Client</title>
  <link rel="icon" type="image/svg+xml" href={logo} />
</svelte:head>

<SvelteToast />

<main>
  <Container fluid class="mb-2">
    <MainHeader />
    <!-- Login Help Modal -->
    <Modal isOpen={showHelpModal} toggle={() => (showHelpModal = false)} theme={$theme}>
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
          <Card outline theme={$theme} color="notifiarr">
            <CardBody class="text-nowrap fs-3">
              <!-- This is before translations are loaded. This 'typo' is on purpose, sue me. -->
              <Spinner /> Translateratating!...</CardBody>
          </Card>
        </Col>
      {:else}
        {#await profile.fetch()}
          <!-- Wait for profile to load. -->
          <Col xs={{ size: 8, offset: 2 }} md={{ size: 4, offset: 4 }}>
            <Card outline theme={$theme} color="notifiarr">
              <CardBody class="text-nowrap fs-3">
                <Spinner /> {$_('phrases.Loading')}</CardBody>
            </Card>
          </Col>
        {:then error}
          {#if $profile.loggedIn}
            <!-- This is the main page, after logging in. -->
            <Navigation />
          {:else}
            <!-- This is the login page, before logging in. -->
            <Col xs={{ size: 8, offset: 2 }} md={{ size: 4, offset: 4 }}>
              <Card outline theme={$theme} class="mt-2" color="notifiarr">
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
                    {#if error}
                      • <span class="text-danger">{error}</span>
                    {/if}
                  </CardFooter>
                </CardBody>
              </Card>
            </Col>
          {/if}
        {/await}
      {/if}
    </Row>
  </Container>
</main>
