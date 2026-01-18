<script lang="ts">
  import { showMsg } from './header/Index.svelte'
  import { delay } from './includes/util'
  import { _ } from './includes/Translate.svelte'
  import { profile } from './api/profile.svelte'
  import {
    Button,
    Spinner,
    CardBody,
    CardFooter,
    Input,
    CardHeader,
    CardTitle,
  } from '@sveltestrap/sveltestrap'
  import Nodal from './includes/Nodal.svelte'

  let username = $state('')
  let password = $state('')
  let isLoading = $state(false)
  let helpModal = $state<Nodal>()
  let { error } = $props()

  async function onsubmit(e: Event) {
    e.preventDefault()
    if (!username || !password) {
      error = $_('config.errors.PleaseEnterBothUsernameAndPassword')
      return
    }

    isLoading = true
    error = ''
    error = (await profile.login(username, password)) ?? ''
    isLoading = false
    if (error) error = $_('config.errors.LoginFailed', { values: { error } })
    else loggedIn()
  }

  const loggedIn = async () => {
    showMsg($_('phrases.LoggedIn'))
    await delay(4567)
    showMsg('')
  }
</script>

<!-- Login Help Modal -->
<Nodal title="phrases.LoginHelp" bind:this={helpModal}>
  {@html $_('phrases.LoginHelpBody')}
</Nodal>

<CardHeader><CardTitle>{$_('buttons.Login')}</CardTitle></CardHeader>
<CardBody>
  <form {onsubmit}>
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
    <Button type="submit" size="sm" disabled={isLoading} class="w-100" color="notifiarr">
      {#if isLoading}<Spinner size="sm" />{/if}
      <span class="fs-5">{$_(isLoading ? 'phrases.LoggingIn' : 'buttons.Login')}</span>
    </Button>
  </form>
</CardBody>

<CardFooter class="mt-2">
  <a href="#showhelp" onclick={helpModal?.open}>{$_('phrases.LoginHelp')}</a>
  {#if error}• <span class="text-danger">{error}</span>{/if}
</CardFooter>
