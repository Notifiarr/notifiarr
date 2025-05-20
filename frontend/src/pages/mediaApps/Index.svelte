<script lang="ts" module>
  import { faClapperboardPlay } from '@fortawesome/sharp-duotone-light-svg-icons'

  export const page = {
    id: 'MediaApps',
    i: faClapperboardPlay,
    c1: 'indigo',
    c2: 'blue',
    d1: 'lightseagreen',
    d2: 'antiquewhite',
  }
</script>

<script lang="ts">
  import { postUi, type BackendResponse } from '../../api/fetch'
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import type { PlexConfig, TautulliConfig } from '../../api/notifiarrConfig'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Instances from '../../includes/Instances.svelte'
  import plexLogo from '../../assets/logos/plex.png'
  import tautulliLogo from '../../assets/logos/tautulli.png'

  let tautulli = $state($profile.config.tautulli)
  let plex = $state($profile.config.plex)
  const conf = $derived($profile.config)
  const api = 'checkInstance/'

  const plexApp = {
    id: page.id + '.Plex',
    logo: plexLogo,
    hidden: ['deletes'],
    disabled: ['name'],
    test: async (plex: PlexConfig, id: number): Promise<BackendResponse> =>
      await postUi(api + 'plex/' + id, JSON.stringify({ ...conf, plex }), false),
  }

  const tautulliApp = {
    id: page.id + '.Tautulli',
    logo: tautulliLogo,
    hidden: ['deletes'],
    test: async (tautulli: TautulliConfig, id: number): Promise<BackendResponse> =>
      await postUi(api + 'tautulli/' + id, JSON.stringify({ ...conf, tautulli }), false),
  }
</script>

<Header {page} />

<CardBody class="pt-0 mt-0">
  <Instances bind:instances={plex} app={plexApp} />
  <Instances bind:instances={tautulli} app={tautulliApp} />
</CardBody>

<Footer submit={() => profile.writeConfig({ ...conf, tautulli, plex })} />
