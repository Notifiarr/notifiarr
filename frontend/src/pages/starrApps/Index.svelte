<script lang="ts" module>
  import { page } from './page.svelte'

  export { page }

  const flt = $derived({
    Sonarr: new FormListTracker(get(profile).config.sonarr, Starr.Sonarr),
    Radarr: new FormListTracker(get(profile).config.radarr, Starr.Radarr),
    Readarr: new FormListTracker(get(profile).config.readarr, Starr.Readarr),
    Lidarr: new FormListTracker(get(profile).config.lidarr, Starr.Lidarr),
    Prowlarr: new FormListTracker(get(profile).config.prowlarr, Starr.Prowlarr),
  })

  export const getTracker = (app: keyof typeof flt) => flt[app]
</script>

<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody, TabContent } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Tab, { getTab } from '../../includes/InstancesTab.svelte'
  import { Starr } from './page.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import type {
    LidarrConfig,
    ProwlarrConfig,
    RadarrConfig,
    ReadarrConfig,
    SonarrConfig,
  } from '../../api/notifiarrConfig'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import { get } from 'svelte/store'

  async function submit() {
    const c = {
      ...$profile.config,
      sonarr: flt.Sonarr.instances as SonarrConfig[],
      radarr: flt.Radarr.instances as RadarrConfig[],
      readarr: flt.Readarr.instances as ReadarrConfig[],
      lidarr: flt.Lidarr.instances as LidarrConfig[],
      prowlarr: flt.Prowlarr.instances as ProwlarrConfig[],
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

{Object.values(flt).every(iv => !iv.formChanged)}
{Object.values(flt).every(iv => iv.removed.length === 0)}
{Object.values(flt).every(iv => iv.invalid)}
<Footer
  {submit}
  saveDisabled={!nav.formChanged && Object.values(flt).some(iv => iv.invalid)} />
