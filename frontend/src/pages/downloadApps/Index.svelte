<script lang="ts" module>
  import { Downloaders as App, page } from './page.svelte'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  export { page }
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
  saveDisabled={!nav.formChanged || Object.values(flt).some(iv => iv.invalid)} />
