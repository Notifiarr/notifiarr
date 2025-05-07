<script lang="ts">
  import {
    Button,
    Card,
    CardBody,
    CardFooter,
    CardHeader,
    Input,
    Navbar,
    NavbarBrand,
    Styles,
  } from '@sveltestrap/sveltestrap'
  import logo from './assets/notifiarr.svg'
  import { profile, fetchProfile } from './api/profile'
  import { login } from './api/login'
  import Navigation from './Navigation.svelte'
  import { SvelteToast } from '@zerodevx/svelte-toast'

  let username = ''
  let password = ''
  let loginFailedMsg = ''
  let isLoading = false

  async function handleLogin() {
    if (!username || !password) {
      loginFailedMsg = 'Please enter both username and password'
      return
    }

    isLoading = true
    loginFailedMsg = ''

    try {
      loginFailedMsg = (await login(username, password)) ?? ''
    } catch (err) {
      loginFailedMsg = `An unexpected error occurred: ${err}`
    } finally {
      isLoading = false
    }
  }
</script>

<svelte:head>
  <title>{$profile.loggedIn ? '' : 'Login - '}Notifiarr Client</title>
  <link rel="icon" type="image/png" href={logo} />
</svelte:head>

<Styles />
<SvelteToast />

<main>
  <Navbar>
    <NavbarBrand href="#">
      <h1><img src={logo} height="60" alt="Notifiarr" /> Notifiarr Client</h1>
    </NavbarBrand>
  </Navbar>

  {#await fetchProfile()}
    <Card body theme="light" color="warning" outline>
      <div>Loading...</div>
    </Card>
  {:then}
    {#if $profile.loggedIn}<!-- This is the main page, after logging in. -->
      <Navigation />
    {:else}<!-- This is the login page, before logging in. -->
      <Card body theme="light" color="info" outline>
        {#if loginFailedMsg}
          <div class="error-message">{loginFailedMsg}</div>
        {/if}
        <form on:submit|preventDefault={handleLogin}>
          <Input
            type="text"
            name="username"
            placeholder="Username"
            bind:value={username} />
          <Input
            type="password"
            name="password"
            placeholder="Password"
            bind:value={password} />
          <Button type="submit" disabled={isLoading}>
            {isLoading ? 'Logging in...' : 'Login'}
          </Button>
        </form>
      </Card>
    {/if}
  {:catch error}
    <!-- error fetching profile (ie. timeout) -->
    <Card body theme="light" color="danger" outline>
      <CardHeader>ERROR</CardHeader>
      <CardBody>{error.message}</CardBody>
      <CardFooter>Try refreshing the page.</CardFooter>
    </Card>
  {/await}
</main>

<style>
  .error-message {
    color: red;
    margin-bottom: 1rem;
  }
</style>
