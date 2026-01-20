<script lang="ts">
  import { showMsg } from './header/Index.svelte'
  import { delay } from './includes/util'
  import T, { _ } from './includes/Translate.svelte'
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

  let apiKey = $state('')
  let isLoading = $state(false)
  let error = $state('')

  async function onsubmit(e: Event) {
    e.preventDefault()
    isLoading = true
    error = ''
    error = (await profile.setApiKey(apiKey)) ?? ''
    if (error) error = $_('config.errors.SetAPIKeyFailed', { values: { error } })
    else {
      showMsg($_('phrases.APIKey.setMsg'))
      $profile.config.apiKey = apiKey
      $profile.loggedIn = false
    }

    await delay(4000) // Wait 4 seconds to prevent spamming the API.
    isLoading = false
  }
</script>

<CardHeader><CardTitle>{$_('phrases.APIKey.title')}</CardTitle></CardHeader>
<CardBody>
  <p><T id="phrases.APIKey.description" /></p>
  <form {onsubmit}>
    <Input
      type="text"
      name="apikey"
      id="apikey"
      placeholder={$_('phrases.APIKey.placeholder')}
      bind:value={apiKey} />
    <Button
      type="submit"
      size="sm"
      disabled={isLoading || apiKey.length !== 36}
      class="w-100 mt-2"
      color="notifiarr">
      {#if isLoading}<Spinner size="sm" />{/if}
      <span class="fs-5">
        {$_(isLoading ? 'phrases.APIKey.Setting' : 'buttons.Save')}
      </span>
    </Button>
  </form>
</CardBody>

{#if error}
  <CardFooter>
    • <span class="text-danger">{error}</span>
  </CardFooter>
{/if}
