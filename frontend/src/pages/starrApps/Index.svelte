<script lang="ts" module>
  import { Starr, page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody, TabContent } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Tab, { getTab } from '../../includes/InstancesTab.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import type { StarrConfig } from '../../api/notifiarrConfig'
  import { FormListTracker } from '../../includes/formsTracker.svelte'

  const flt = $derived({
    Sonarr: new FormListTracker($profile.config.sonarr ?? [], Starr.Sonarr),
    Radarr: new FormListTracker($profile.config.radarr ?? [], Starr.Radarr),
    Readarr: new FormListTracker($profile.config.readarr ?? [], Starr.Readarr),
    Lidarr: new FormListTracker($profile.config.lidarr ?? [], Starr.Lidarr),
    Prowlarr: new FormListTracker($profile.config.prowlarr ?? [], Starr.Prowlarr),
  })

  async function submit() {
    const c = {
      ...$profile.config,
      sonarr: flt.Sonarr.instances as StarrConfig[],
      radarr: flt.Radarr.instances as StarrConfig[],
      readarr: flt.Readarr.instances as StarrConfig[],
      lidarr: flt.Lidarr.instances as StarrConfig[],
      prowlarr: flt.Prowlarr.instances as StarrConfig[],
    }
    await profile.writeConfig(c)

    if (profile.error) return
    // clears the delete counters.
    Object.values(flt).forEach(iv => iv.resetAll())
  }

  let tab = $state(getTab(Starr.tabs))

  $effect(() => {
    nav.formChanged = Object.values(flt).some(iv => iv.formChanged)
  })
</script>

<Header {page} />

<CardBody>
  <TabContent on:tab={e => nav.goto(e, page.id, [e.detail.toString()])}>
    <Tab flt={flt.Sonarr} bind:tab titles={Starr.title} />
    <Tab flt={flt.Radarr} bind:tab titles={Starr.title} />
    <Tab flt={flt.Readarr} bind:tab titles={Starr.title} />
    <Tab flt={flt.Lidarr} bind:tab titles={Starr.title} />
    <Tab flt={flt.Prowlarr} bind:tab titles={Starr.title} />
  </TabContent>
</CardBody>

<Footer
  {submit}
  saveDisabled={!nav.formChanged || Object.values(flt).some(iv => iv.invalid)} />
