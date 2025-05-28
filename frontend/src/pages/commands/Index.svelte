<script lang="ts" module>
  import { app, page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import type { Command } from '../../api/notifiarrConfig'
  import { FormListTracker, type App } from '../../includes/formsTracker.svelte'
  import { nav } from '../../navigation/nav.svelte'

  let flt = $derived(new FormListTracker($profile.config.commands ?? [], app))

  $effect(() => {
    nav.formChanged = flt.formChanged
  })

  const submit = async () => {
    await profile.writeConfig({
      ...$profile.config,
      commands: flt.instances as Command[],
    })
    if (!profile.error) flt.resetAll() // clears the delete counters.
  }
</script>

<Header {page} />

<CardBody class="pt-0 mt-0"></CardBody>

<Footer {submit} saveDisabled={!flt.formChanged || flt.invalid} />
