<script lang="ts" module>
  import { app, page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody, Col, Row } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import Command from './Command.svelte'
  import Instances from '../../includes/Instances.svelte'
  import TestRegex from '../../includes/TestRegex.svelte'

  let flt = $derived(new FormListTracker($profile.config.commands ?? [], app))

  $effect(() => {
    nav.formChanged = flt.formChanged
  })

  const submit = async () => {
    await profile.writeConfig({ ...$profile.config, commands: flt.instances })
    if (!profile.error) flt.resetAll() // clears the delete counters.
  }
</script>

<Header {page} />

<CardBody>
  <Instances {flt} Child={Command} deleteButton={page.id + '.DeleteCommand'}>
    {#snippet headerActive(index)}
      {index + 1}. {flt.original?.[index]?.name}
    {/snippet}
    {#snippet headerCollapsed(index)}
      {flt.original?.[index]?.command}
    {/snippet}
  </Instances>

  <!-- Test regular expression -->
  <Row><Col><TestRegex /></Col></Row>
</CardBody>

<Footer {submit} saveDisabled={!flt.formChanged || flt.invalid} />
