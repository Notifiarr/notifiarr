<script lang="ts" module>
  import { page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import { app } from './page.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import Instances from '../../includes/Instances.svelte'
  import Endpoint from './Endpoint.svelte'

  let flt = $derived(new FormListTracker($profile.config.endpoints ?? [], app))

  $effect(() => {
    nav.formChanged = flt.formChanged
  })

  const submit = async () => {
    await profile.writeConfig({ ...$profile.config, endpoints: flt.instances })
    if (!profile.error) flt.resetAll() // clears the delete counters.
  }
</script>

<Header {page} />

<CardBody>
  <Instances {flt} Child={Endpoint} deleteButton={page.id + '.DeleteEndpoint'}>
    {#snippet headerActive(index)}
      {index + 1}. {flt.original?.[index]?.name}
    {/snippet}
    {#snippet headerCollapsed(index)}
      {flt.original?.[index]?.url}
    {/snippet}
  </Instances>
</CardBody>

<Footer {submit} saveDisabled={!flt.formChanged || flt.invalid} />
