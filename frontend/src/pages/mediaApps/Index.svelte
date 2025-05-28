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
  import Instance from '../../includes/Instance.svelte'
  import InstanceHeader from '../../includes/InstanceHeader.svelte'
  import type { Config, PlexConfig, TautulliConfig } from '../../api/notifiarrConfig'
  import { FormListTracker, type App } from '../../includes/formsTracker.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import { validate } from '../../includes/instanceValidator'

  const plexApp: App<PlexConfig> = {
    name: 'Plex',
    id: page.id + '.Plex',
    logo: plexLogo,
    hidden: ['deletes'],
    disabled: ['name'],
    merge: (index: number, form: PlexConfig) => ({
      ...({} as Config),
      plex: form as PlexConfig,
    }),
    // Ignore empty values since Plex is optional.
    validator: (id: string, value: any, index: number) => {
      if (id.endsWith('.name')) return ''
      if (id.endsWith('.url') && value === '') return ''
      if (id.endsWith('.token') && value === '') return ''
      return validate(id, value, index, [$profile.config.plex ?? {}])
    },
  }

  const tautulliApp: App<TautulliConfig> = {
    name: 'Tautulli',
    id: page.id + '.Tautulli',
    logo: tautulliLogo,
    hidden: ['deletes'],
    merge: (index: number, form: TautulliConfig) => ({
      ...({} as Config),
      tautulli: form as TautulliConfig,
    }),
    // Ignore empty values since Plex is optional.
    validator: (id: string, value: any, index: number) => {
      if (id.endsWith('.name')) return ''
      if (id.endsWith('.url') && value === '') return ''
      if (id.endsWith('.apiKey') && value === '') return ''
      return validate(id, value, index, [$profile.config.tautulli ?? {}])
    },
  }

  let iv = $derived({
    Plex: new FormListTracker([$profile.config.plex ?? ({} as PlexConfig)], plexApp),
    Tautulli: new FormListTracker(
      [$profile.config.tautulli ?? ({} as TautulliConfig)],
      tautulliApp,
    ),
  })

  $effect(() => {
    nav.formChanged = Object.values(iv).some(iv => iv.formChanged)
  })
</script>

<Header {page} />

<!-- We use the zero index because we only support once of each of these apps. -->
<CardBody class="pt-0 mt-0">
  <InstanceHeader flt={iv.Plex} />
  <Instance
    reset={() => iv.Plex.resetForm(0)}
    validate={(id, value) => iv.Plex.validate(id, value, 0)}
    bind:form={iv.Plex.instances[0]!}
    original={iv.Plex.original[0]!}
    app={plexApp} />
  <InstanceHeader flt={iv.Tautulli} />
  <Instance
    reset={() => iv.Tautulli.resetForm(0)}
    validate={(id, value) => iv.Tautulli.validate(id, value, 0)}
    bind:form={iv.Tautulli.instances[0]!}
    original={iv.Tautulli.original[0]!}
    app={tautulliApp} />
</CardBody>

<Footer
  submit={() =>
    profile.writeConfig({
      ...$profile.config,
      plex: iv.Plex.instances[0]! as PlexConfig,
      tautulli: iv.Tautulli.instances[0]! as TautulliConfig,
    })}
  saveDisabled={!nav.formChanged || Object.values(iv).some(iv => iv.invalid)} />
