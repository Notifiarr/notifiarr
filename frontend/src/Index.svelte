<script lang="ts">
  import {
    Card,
    CardBody,
    CardHeader,
    CardTitle,
    Col,
    Container,
    Row,
    Spinner,
  } from '@sveltestrap/sveltestrap'
  import logo from './assets/notifiarr.svg?inline'
  import { profile } from './api/profile.svelte'
  import Navigation from './navigation/Index.svelte'
  import { SvelteToast } from '@zerodevx/svelte-toast'
  import { isReady, _ } from './includes/Translate.svelte'
  import { setLocale } from './includes/locale/index.svelte'
  import { onMount } from 'svelte'
  import { theme } from './includes/theme.svelte'
  import MainHeader from './header/Index.svelte'
  import Login from './Login.svelte'

  onMount(() => {
    const query = new URLSearchParams(window.location.search)
    if (query.get('lang')) setLocale(query.get('lang')!)
  })
</script>

<svelte:head>
  <title>{$profile.loggedIn ? '' : `Login - `}Notifiarr Client</title>
  <link rel="icon" type="image/svg+xml" href={logo} />
</svelte:head>

<SvelteToast />

<main>
  <Container fluid class="mb-2">
    <MainHeader />

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
            <Col xs={{ size: 8, offset: 2 }} md={{ size: 4, offset: 4 }}>
              <!-- This is the login page, before logging in. -->
              <Card outline theme={$theme} class="mt-2" color="notifiarr">
                <CardHeader><CardTitle>{$_('buttons.Login')}</CardTitle></CardHeader>
                <Login {error} />
              </Card>
            </Col>
          {/if}
        {/await}
      {/if}
    </Row>
  </Container>
</main>
