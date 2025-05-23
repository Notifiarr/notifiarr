<script lang="ts" module>
  import { page } from './starr.svelte'
  export { page }
</script>

<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody, TabContent } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Instances from '../../includes/Instances.svelte'
  import { Starr } from './starr.svelte'
  import { deepCopy, deepEqual } from '../../includes/util'
  import type { Form } from '../../includes/Instance.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import Tab from './Tab.svelte'

  const validate = (id: string, index: number, value: any, reset?: boolean) => {
    const app = id.split('.')[1]

    if (!invalid[app]) invalid[app] = {}
    if (!invalid[app][index]) invalid[app][index] = {}

    if (reset) {
      Object.keys(invalid[app][index]).forEach(id => (invalid[app][index][id] = ''))
      return ''
    }

    // Find the right instances array from the input id.
    let appFn: (Form | null)[] | null = null
    switch (app) {
      case 'Radarr':
        appFn = radarr
        break
      case 'Readarr':
        appFn = readarr
        break
      case 'Lidarr':
        appFn = lidarr
        break
      case 'Prowlarr':
        appFn = prowlarr
        break
      default:
        appFn = sonarr
        break
    }

    // Call the primary validate function (from another file).
    invalid[app][index][id] = Starr.Validate(id, index, value, appFn)
    return invalid[app][index][id]
  }

  // These are the form bindings.
  // These do not change.
  let sonarr = $state(deepCopy($profile.config.sonarr ?? []))
  let radarr = $state(deepCopy($profile.config.radarr ?? []))
  let readarr = $state(deepCopy($profile.config.readarr ?? []))
  let lidarr = $state(deepCopy($profile.config.lidarr ?? []))
  let prowlarr = $state(deepCopy($profile.config.prowlarr ?? []))

  // Original values so we can compare them for changes.
  // These could change if they backend is updated.
  let originalSonarr = $derived(deepCopy($profile.config.sonarr ?? []))
  let originalRadarr = $derived(deepCopy($profile.config.radarr ?? []))
  let originalReadarr = $derived(deepCopy($profile.config.readarr ?? []))
  let originalLidarr = $derived(deepCopy($profile.config.lidarr ?? []))
  let originalProwlarr = $derived(deepCopy($profile.config.prowlarr ?? []))

  // Allows binding to the instances component so we can clear the delete counters.
  let iSonarr = $state<Instances>()
  let iRadarr = $state<Instances>()
  let iReadarr = $state<Instances>()
  let iLidarr = $state<Instances>()
  let iProwlarr = $state<Instances>()

  let removed = $state({ Sonarr: [], Radarr: [], Readarr: [], Lidarr: [], Prowlarr: [] })

  let formChanged = $derived(
    !deepEqual(sonarr, originalSonarr) ||
      !deepEqual(radarr, originalRadarr) ||
      !deepEqual(readarr, originalReadarr) ||
      !deepEqual(lidarr, originalLidarr) ||
      !deepEqual(prowlarr, originalProwlarr),
  )

  async function submit() {
    const c = { ...$profile.config, sonarr, radarr, readarr, lidarr, prowlarr }
    await profile.writeConfig(c)
    // clears the delete counters.
    iSonarr?.clear()
    iRadarr?.clear()
    iReadarr?.clear()
    iLidarr?.clear()
    iProwlarr?.clear()
  }

  // Keep track of invalid states and form changes.
  let invalid = $state<Record<string, Record<number, Record<string, string>>>>({})
  const allValid = $derived(
    Object.values(invalid).every(v =>
      Object.values(v).every(v => Object.values(v).every(v => !v)),
    ),
  )

  const sonarrInvalid = $derived(
    Object.values(invalid[Starr.title.Sonarr] ?? {}).some(v =>
      Object.values(v).some(v => !!v),
    ),
  )
  const radarrInvalid = $derived(
    Object.values(invalid[Starr.title.Radarr] ?? {}).some(v =>
      Object.values(v).some(v => !!v),
    ),
  )
  const readarrInvalid = $derived(
    Object.values(invalid[Starr.title.Readarr] ?? {}).some(v =>
      Object.values(v).some(v => !!v),
    ),
  )
  const lidarrInvalid = $derived(
    Object.values(invalid[Starr.title.Lidarr] ?? {}).some(v =>
      Object.values(v).some(v => !!v),
    ),
  )
  const prowlarrInvalid = $derived(
    Object.values(invalid[Starr.title.Prowlarr] ?? {}).some(v =>
      Object.values(v).some(v => !!v),
    ),
  )

  const tabs = ['sonarr', 'radarr', 'readarr', 'lidarr', 'prowlarr']
  const uriTab = window.location.pathname.split('/').pop() ?? tabs[0]
  let tab = $state(Object.values(tabs).includes(uriTab) ? uriTab : tabs[0])
</script>

<Header {page} />

<CardBody>
  <TabContent on:tab={e => nav.goto(e, page.id, [e.detail.toString()])}>
    <Tab
      app={Starr.Sonarr}
      equal={deepEqual(sonarr, originalSonarr)}
      original={originalSonarr}
      valid={!sonarrInvalid}
      {validate}
      bind:form={iSonarr}
      bind:instances={sonarr}
      bind:removed
      bind:tab />

    <Tab
      app={Starr.Radarr}
      equal={deepEqual(radarr, originalRadarr)}
      original={originalRadarr}
      valid={!radarrInvalid}
      {validate}
      bind:form={iRadarr}
      bind:instances={radarr}
      bind:removed
      bind:tab />

    <Tab
      app={Starr.Readarr}
      equal={deepEqual(readarr, originalReadarr)}
      original={originalReadarr}
      valid={!readarrInvalid}
      {validate}
      bind:form={iReadarr}
      bind:instances={readarr}
      bind:removed
      bind:tab />

    <Tab
      app={Starr.Lidarr}
      equal={deepEqual(lidarr, originalLidarr)}
      original={originalLidarr}
      valid={!lidarrInvalid}
      {validate}
      bind:form={iLidarr}
      bind:instances={lidarr}
      bind:removed
      bind:tab />

    <Tab
      app={Starr.Prowlarr}
      equal={deepEqual(prowlarr, originalProwlarr)}
      original={originalProwlarr}
      valid={!prowlarrInvalid}
      {validate}
      bind:form={iProwlarr}
      bind:instances={prowlarr}
      bind:removed
      bind:tab />
  </TabContent>
</CardBody>

<Footer
  {submit}
  saveDisabled={(!formChanged &&
    removed.Sonarr.length === 0 &&
    removed.Radarr.length === 0 &&
    removed.Readarr.length === 0 &&
    removed.Lidarr.length === 0 &&
    removed.Prowlarr.length === 0) ||
    !allValid} />
