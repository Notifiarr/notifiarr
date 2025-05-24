<script lang="ts" module>
  import { Downloaders as App, page } from './page.svelte'
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
  import { InstanceFormValidator } from '../../includes/instanceFormValidator.svelte'
  import type {
    DelugeConfig,
    NZBGetConfig,
    QbitConfig,
    RtorrentConfig,
    SabNZBConfig,
    XmissionConfig,
  } from '../../api/notifiarrConfig'

  let iv = $derived({
    Qbittorrent: new InstanceFormValidator($profile.config.qbit, App.Qbittorrent),
    Rtorrent: new InstanceFormValidator($profile.config.rtorrent, App.Rtorrent),
    Transmission: new InstanceFormValidator($profile.config.transmission, App.Xmission),
    Deluge: new InstanceFormValidator($profile.config.deluge, App.Deluge),
    SabNZB: new InstanceFormValidator($profile.config.sabnzbd, App.SabNZB),
    NZBGet: new InstanceFormValidator($profile.config.nzbget, App.NZBGet),
  })

  async function submit() {
    const c = {
      ...$profile.config,
      qbit: iv.Qbittorrent.instances as QbitConfig[],
      rtorrent: iv.Rtorrent.instances as RtorrentConfig[],
      transmission: iv.Transmission.instances as XmissionConfig[],
      deluge: iv.Deluge.instances as DelugeConfig[],
      sabnzbd: iv.SabNZB.instances as SabNZBConfig[],
      nzbget: iv.NZBGet.instances as NZBGetConfig[],
    }
    await profile.writeConfig(c)

    if (profile.error) return
    // clears the delete counters.
    Object.values(iv).forEach(iv => iv.resetAll())
  }

  let tab = $state(getTab(App.tabs))

  $effect(() => {
    nav.formChanged = Object.values(iv).some(iv => iv.formChanged)
  })
</script>

<Header {page} />
<CardBody>
  <TabContent on:tab={e => nav.goto(e, page.id, [e.detail.toString()])}>
    <Tab iv={iv.Qbittorrent} bind:tab titles={App.title} />
    <Tab iv={iv.Rtorrent} bind:tab titles={App.title} />
    <Tab iv={iv.Transmission} bind:tab titles={App.title} />
    <Tab iv={iv.Deluge} bind:tab titles={App.title} />
    <Tab iv={iv.SabNZB} bind:tab titles={App.title} />
    <Tab iv={iv.NZBGet} bind:tab titles={App.title} />
  </TabContent>
</CardBody>

<Footer
  {submit}
  saveDisabled={(!nav.formChanged &&
    Object.values(iv).every(iv => iv.removed.length === 0)) ||
    Object.values(iv).some(iv => iv.invalid)} />
