<script lang="ts" module>
  import { page } from './page.svelte'
  export { page }
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
  import { InstanceFormValidator } from '../../includes/instanceFormValidator.svelte'
  import type {
    LidarrConfig,
    ProwlarrConfig,
    RadarrConfig,
    ReadarrConfig,
    SonarrConfig,
  } from '../../api/notifiarrConfig'

  let iv = $derived({
    Sonarr: new InstanceFormValidator($profile.config.sonarr, Starr.Sonarr),
    Radarr: new InstanceFormValidator($profile.config.radarr, Starr.Radarr),
    Readarr: new InstanceFormValidator($profile.config.readarr, Starr.Readarr),
    Lidarr: new InstanceFormValidator($profile.config.lidarr, Starr.Lidarr),
    Prowlarr: new InstanceFormValidator($profile.config.prowlarr, Starr.Prowlarr),
  })

  async function submit() {
    const c = {
      ...$profile.config,
      sonarr: iv.Sonarr.instances as SonarrConfig[],
      radarr: iv.Radarr.instances as RadarrConfig[],
      readarr: iv.Readarr.instances as ReadarrConfig[],
      lidarr: iv.Lidarr.instances as LidarrConfig[],
      prowlarr: iv.Prowlarr.instances as ProwlarrConfig[],
    }
    await profile.writeConfig(c)

    if (profile.error) return
    // clears the delete counters.
    Object.values(iv).forEach(iv => iv.resetAll())
  }

  let tab = $state(getTab(Starr.tabs))

  $effect(() => {
    nav.formChanged = Object.values(iv).some(iv => iv.formChanged)
  })
</script>

<Header {page} />

<CardBody>
  <TabContent on:tab={e => nav.goto(e, page.id, [e.detail.toString()])}>
    <Tab iv={iv.Sonarr} bind:tab titles={Starr.title} />
    <Tab iv={iv.Radarr} bind:tab titles={Starr.title} />
    <Tab iv={iv.Readarr} bind:tab titles={Starr.title} />
    <Tab iv={iv.Lidarr} bind:tab titles={Starr.title} />
    <Tab iv={iv.Prowlarr} bind:tab titles={Starr.title} />
  </TabContent>
</CardBody>

<Footer
  {submit}
  saveDisabled={(Object.values(iv).every(iv => !iv.formChanged) &&
    Object.values(iv).every(iv => iv.removed.length === 0)) ||
    !Object.values(iv).every(iv => iv.invalid)} />
