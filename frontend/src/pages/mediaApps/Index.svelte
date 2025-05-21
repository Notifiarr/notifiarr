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
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Instances from '../../includes/Instances.svelte'
  import plexLogo from '../../assets/logos/plex.png'
  import tautulliLogo from '../../assets/logos/tautulli.png'

  let tautulli = $state($profile.config.tautulli)
  let plex = $state($profile.config.plex)
  const conf = $derived($profile.config)

  const plexApp = {
    name: 'Plex',
    id: page.id + '.Plex',
    logo: plexLogo,
    hidden: ['deletes'],
    disabled: ['name'],
  }

  const tautulliApp = {
    name: 'Tautulli',
    id: page.id + '.Tautulli',
    logo: tautulliLogo,
    hidden: ['deletes'],
  }
</script>

<Header {page} />

<CardBody class="pt-0 mt-0">
  <Instances bind:instances={plex} app={plexApp} />
  <Instances bind:instances={tautulli} app={tautulliApp} />
</CardBody>

<Footer submit={() => profile.writeConfig({ ...conf, tautulli, plex })} />
