<script lang="ts" module>
  import { get } from 'svelte/store'
  import { Downloaders as App, page } from './page.svelte'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  export { page }

  let flt = $derived({
    Qbittorrent: new FormListTracker(get(profile).config.qbit, App.Qbittorrent),
    Rtorrent: new FormListTracker(get(profile).config.rtorrent, App.Rtorrent),
    Transmission: new FormListTracker(get(profile).config.transmission, App.Xmission),
    Deluge: new FormListTracker(get(profile).config.deluge, App.Deluge),
    SabNZB: new FormListTracker(get(profile).config.sabnzbd, App.SabNZB),
    NZBGet: new FormListTracker(get(profile).config.nzbget, App.NZBGet),
  })
  export const getTracker = (app: keyof typeof flt) => flt[app]
</script>

<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody, TabContent } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import Tab, { getTab } from '../../includes/InstancesTab.svelte'
  import type {
    DelugeConfig,
    NZBGetConfig,
    QbitConfig,
    RtorrentConfig,
    SabNZBConfig,
    XmissionConfig,
  } from '../../api/notifiarrConfig'

  async function submit() {
    const c = {
      ...$profile.config,
      qbit: flt.Qbittorrent.instances as QbitConfig[],
      rtorrent: flt.Rtorrent.instances as RtorrentConfig[],
      transmission: flt.Transmission.instances as XmissionConfig[],
      deluge: flt.Deluge.instances as DelugeConfig[],
      sabnzbd: flt.SabNZB.instances as SabNZBConfig[],
      nzbget: flt.NZBGet.instances as NZBGetConfig[],
    }
    await profile.writeConfig(c)

    if (profile.error) return
    // clears the delete counters.
    Object.values(flt).forEach(iv => iv.resetAll())
  }

  let tab = $state(getTab(App.tabs))

  $effect(() => {
    nav.formChanged = Object.values(flt).some(iv => iv.formChanged)
  })
</script>

<Header {page} />
<CardBody>
  <TabContent on:tab={e => nav.goto(e, page.id, [e.detail.toString()])}>
    <Tab flt={flt.Qbittorrent} bind:tab titles={App.title} />
    <Tab flt={flt.Rtorrent} bind:tab titles={App.title} />
    <Tab flt={flt.Transmission} bind:tab titles={App.title} />
    <Tab flt={flt.Deluge} bind:tab titles={App.title} />
    <Tab flt={flt.SabNZB} bind:tab titles={App.title} />
    <Tab flt={flt.NZBGet} bind:tab titles={App.title} />
  </TabContent>
</CardBody>

<Footer
  {submit}
  saveDisabled={!nav.formChanged && Object.values(flt).some(iv => iv.invalid)} />
