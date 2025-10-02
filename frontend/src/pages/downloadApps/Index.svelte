<script lang="ts" module>
  import { Downloaders as App, page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody, TabContent } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import Tab, { setTab, goto } from '../../includes/InstancesTab.svelte'
  import type {
    DelugeConfig,
    NZBGetConfig,
    QbitConfig,
    RtorrentConfig,
    SabNZBConfig,
    XmissionConfig,
  } from '../../api/notifiarrConfig'

  let flt = $derived({
    Qbittorrent: new FormListTracker($profile.config.qbit ?? [], App.Qbittorrent),
    Rtorrent: new FormListTracker($profile.config.rtorrent ?? [], App.Rtorrent),
    Transmission: new FormListTracker($profile.config.transmission ?? [], App.Xmission),
    Deluge: new FormListTracker($profile.config.deluge ?? [], App.Deluge),
    SabNZB: new FormListTracker($profile.config.sabnzbd ?? [], App.SabNZB),
    NZBGet: new FormListTracker($profile.config.nzbget ?? [], App.NZBGet),
  })

  async function submit() {
    const c = { ...$profile.config }
    c.qbit = flt.Qbittorrent.instances as QbitConfig[]
    c.rtorrent = flt.Rtorrent.instances as RtorrentConfig[]
    c.transmission = flt.Transmission.instances as XmissionConfig[]
    c.deluge = flt.Deluge.instances as DelugeConfig[]
    c.sabnzbd = flt.SabNZB.instances as SabNZBConfig[]
    c.nzbget = flt.NZBGet.instances as NZBGetConfig[]
    await profile.writeConfig(c)
    if (!profile.error) Object.values(flt).forEach(iv => iv.resetAll())
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
  <!-- only nav.goto if the tab is different -->
  <TabContent on:tab={e => goto(e, page.id)}>
    <Tab flt={flt.Qbittorrent} titles={App.title} />
    <Tab flt={flt.Rtorrent} titles={App.title} />
    <Tab flt={flt.Transmission} titles={App.title} />
    <Tab flt={flt.Deluge} titles={App.title} />
    <Tab flt={flt.SabNZB} titles={App.title} />
    <Tab flt={flt.NZBGet} titles={App.title} />
  </TabContent>
</CardBody>

<Footer
  {submit}
  saveDisabled={!nav.formChanged || Object.values(flt).some(iv => iv.invalid)} />
