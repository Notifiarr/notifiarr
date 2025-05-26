<script lang="ts" module>
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import { app, page } from './page.svelte'
  export { page }

  const flt = $derived(new FormListTracker(get(profile).config.watchFiles, app))
</script>

<script lang="ts">
  import { Card, CardBody, CardHeader } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Watcher from './Watcher.svelte'
  import Instances from '../../includes/Instances.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import { profile } from '../../api/profile.svelte'
  import { get } from 'svelte/store'
  import Input from '../../includes/Input.svelte'
  import { slide } from 'svelte/transition'
  import { escapeHtml } from '../../includes/util'
  import TestRegex from './TestRegex.svelte'

  // Handle form submission
  const submit = async () => {
    await profile.writeConfig({ ...$profile.config, watchFiles: flt.instances })
    if (!profile.error) flt.resetAll() // clears the delete counters.
  }

  $effect(() => {
    nav.formChanged = flt.formChanged
  })
</script>

<Header {page} />

<CardBody style="max-width: 100%;">
  <Instances {flt} Child={Watcher} deleteButton={page.id + '.DeleteWatcher'}>
    {#snippet headerActive(index)}
      {index + 1}. {flt.original[index]?.path}
    {/snippet}
    {#snippet headerCollapsed(index)}
      {flt.original[index]?.regex}
    {/snippet}
  </Instances>

  <!-- Test regular expression -->
  <TestRegex />
</CardBody>

<Footer {submit} saveDisabled={!flt.formChanged || flt.invalid} />
