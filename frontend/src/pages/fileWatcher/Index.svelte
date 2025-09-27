<script lang="ts" module>
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import { app, page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import { CardBody, Col, Row } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Watcher from './Watcher.svelte'
  import Instances from '../../includes/Instances.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import { profile } from '../../api/profile.svelte'
  import TestRegex from '../../includes/TestRegex.svelte'

  const flt = $derived(new FormListTracker($profile.config.watchFiles ?? [], app))

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
      {index + 1}. {flt.original?.[index]?.path}
    {/snippet}
    {#snippet headerCollapsed(index)}
      <span class="text-primary">Match:</span>
      {flt.original?.[index]?.regex}
      <span class="text-warning">Skip:</span>
      {flt.original?.[index]?.skip}
    {/snippet}
  </Instances>
  <!-- Test regular expression -->
  <Row><Col><TestRegex /></Col></Row>
</CardBody>

<Footer {submit} saveDisabled={!flt.formChanged || flt.invalid} />
