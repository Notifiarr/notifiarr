<script lang="ts" module>
  import { SnapshotApps as App, page } from './page.svelte'
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody, TabContent } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import Tab, { goto, setTab } from '../../includes/InstancesTab.svelte'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  export { page } // Pass it through.
</script>

<script lang="ts">
  if (!$profile.config.snapshot.nvidia.busIDs?.length) {
    $profile.config.snapshot.nvidia.busIDs = ['']
  }

  let flt = $derived({
    MySQL: new FormListTracker($profile.config.snapshot.mysql ?? [], App.mysqlApp),
    Nvidia: new FormListTracker([$profile.config.snapshot.nvidia], App.nvidiaApp),
  })

  async function submit() {
    const c = { ...$profile.config }
    c.snapshot.mysql = flt.MySQL.instances
    c.snapshot.nvidia = flt.Nvidia.instances[0]
    await profile.writeConfig(c)
  }

  $effect(() => {
    nav.formChanged = Object.values(flt).some(iv => iv.formChanged)
  })
  setTab(App.tabs) // this sets the initial tab.
</script>

<!-- update the tab when the user navigates back -->
<svelte:window on:popstate={() => setTab(App.tabs)} />

<Header {page} />

<CardBody>
  <TabContent on:tab={e => goto(e, page.id)}>
    <Tab bind:flt={flt.MySQL} titles={App.title} />
    <Tab bind:flt={flt.Nvidia} titles={App.title} one />
  </TabContent>
</CardBody>

<Footer
  {submit}
  saveDisabled={!nav.formChanged || Object.values(flt).some(iv => iv.invalid)} />
