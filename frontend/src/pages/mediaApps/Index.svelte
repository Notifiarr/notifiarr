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
  import Instance, { type Form } from '../../includes/Instance.svelte'
  import InstanceHeader from '../../includes/InstanceHeader.svelte'
  import type { Config, PlexConfig, TautulliConfig } from '../../api/notifiarrConfig'
  import { InstanceFormValidator } from '../../includes/instanceFormValidator.svelte'
  import { nav } from '../../navigation/nav.svelte'

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
    // Ignore empty values since Plex is optional.
    customValidator: (id: string, value: any) => (value === '' ? '' : undefined),
    merge,
  }

  const tautulliApp = {
    name: 'Tautulli',
    id: page.id + '.Tautulli',
    logo: tautulliLogo,
    hidden: ['deletes'],
    customValidator: (id: string, value: any) => (value === '' ? '' : undefined),
    merge,
  }

  let iv = $derived({
    Plex: new InstanceFormValidator([$profile.config.plex ?? {}], plexApp),
    Tautulli: new InstanceFormValidator([$profile.config.tautulli ?? {}], tautulliApp),
  })

  $effect(() => {
    nav.formChanged = Object.values(iv).some(iv => iv.formChanged)
  })
</script>

<Header {page} />

<!-- We use the zero index because we only support once of each of these apps. -->
<CardBody class="pt-0 mt-0">
  <InstanceHeader iv={iv.Plex} />
  <Instance
    reset={() => iv.Plex.resetForm(0)}
    validate={(id, value) => iv.Plex.validate(id, value, 0)}
    bind:form={iv.Plex.instances[0]!}
    original={iv.Plex.original[0]!}
    app={plexApp} />
  <InstanceHeader iv={iv.Tautulli} />
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
  saveDisabled={Object.values(iv).every(iv => !iv.formChanged) ||
    Object.values(iv).some(iv => iv.invalid)} />
