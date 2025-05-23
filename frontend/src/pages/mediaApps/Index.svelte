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
  import plexLogo from '../../assets/logos/plex.png'
  import tautulliLogo from '../../assets/logos/tautulli.png'
  import { deepCopy, deepEqual } from '../../includes/util'
  import Instance, { type Form } from '../../includes/Instance.svelte'
  import InstanceHeader from '../../includes/InstanceHeader.svelte'
  import type { Config } from '../../api/notifiarrConfig'

  let tautulli = $state($profile.config.tautulli)
  let plex = $state($profile.config.plex)
  let originalTautulli = $derived(deepCopy($profile.config.tautulli))
  let originalPlex = $derived(deepCopy($profile.config.plex))
  let invalid = $state<Record<number, Record<string, string>>>({})
  let formChanged = $derived(
    !deepEqual(plex, originalPlex) || !deepEqual(tautulli, originalTautulli),
  )

  const validate = (id: string, index: number, value: any) => {
    if (!invalid[index]) invalid[index] = {}

    if (id.endsWith('.url')) {
      invalid[index][id] =
        value.startsWith('http://') || value.startsWith('https://') || value === ''
          ? ''
          : $_('phrases.URLMustBeginWithHttp')
    }
    return invalid[index][id]
  }

  const merge = (index: number, form: Form) => {
    return {
      ...({} as Config),
      // ugly but it works
      [plexApp.name.toLowerCase()]: form,
      [tautulliApp.name.toLowerCase()]: form,
    }
  }

  const plexApp = {
    name: 'Plex',
    id: page.id + '.Plex',
    logo: plexLogo,
    hidden: ['deletes'],
    disabled: ['name'],
    validate,
    merge,
  }

  const tautulliApp = {
    name: 'Tautulli',
    id: page.id + '.Tautulli',
    logo: tautulliLogo,
    hidden: ['deletes'],
    validate,
    merge,
  }

  const allValid = $derived(
    Object.values(invalid).every(v => Object.values(v).every(v => !v)),
  )
</script>

<Header {page} />

<CardBody class="pt-0 mt-0">
  <InstanceHeader app={plexApp} changed={!deepEqual(plex, originalPlex)} />
  <Instance bind:form={plex!} original={originalPlex!} app={plexApp} />
  <InstanceHeader app={tautulliApp} changed={!deepEqual(tautulli, originalTautulli)} />
  <Instance bind:form={tautulli!} original={originalTautulli!} app={tautulliApp} />
</CardBody>

<Footer
  saveDisabled={!formChanged || !allValid}
  submit={() => profile.writeConfig({ ...$profile.config, tautulli, plex })} />
