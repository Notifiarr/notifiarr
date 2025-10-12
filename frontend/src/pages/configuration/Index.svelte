<script lang="ts" module>
  import { faFlaskGear } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'Configuration',
    i: faFlaskGear,
    c1: 'slategray',
    c2: 'darkgreen',
    d1: 'gainsboro',
    d2: 'lime',
  }
</script>

<script lang="ts">
  import { CardBody } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Input from '../../includes/Input.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Logging from './Logging.svelte'
  import Services from './Services.svelte'
  import SslConfig from './SSLConfig.svelte'
  import System from './System.svelte'
  import Network from './Network.svelte'
  import { deepEqual } from '../../includes/util'
  import { nav } from '../../navigation/nav.svelte'
  import Nodal from '../../includes/Nodal.svelte'

  // Local state that syncs with profile store.
  let config = $state($profile.config)
  // Convert array to newline-separated string for textarea
  let extraKeys = $derived($profile.config.extraKeys?.join('\n') ?? '')
  const rows = $derived(
    extraKeys.split('\n').length > 10 ? 10 : extraKeys.split('\n').length,
  )

  // Handle form submission
  function submit() {
    config.extraKeys = extraKeys.split(/\s+/).filter(key => key.trim() !== '')
    profile.writeConfig(config).then(ok => reset(ok))
  }

  const reset = (ok: boolean) => {
    if (ok) config = $profile.config
  }

  // Reset config when profile is updated (when reload button is clicked).
  $effect(() => reset(!!profile.updated))

  // Keep track of form changes. This causes an "unsaved changes" modal to be shown.
  $effect(() => {
    nav.formChanged =
      !deepEqual($profile.config, config) ||
      extraKeys !== ($profile.config.extraKeys?.join('\n') ?? '')
  })

  let isOpen = $state(false)
</script>

<Header
  page={{ ...page, href: '#rawConfig', onclick: () => (isOpen = true) }}
  badge={$_('phrases.Version', { values: { version: config.version } })} />

<CardBody class="pt-0 mt-0">
  <!-- General Section -->
  <h4>{$_('config.titles.General')}</h4>
  <Input
    id="config.apiKey"
    type="password"
    envVar="API_KEY"
    bind:value={config.apiKey}
    original={$profile.config.apiKey} />
  <Input
    id="config.extraKeys"
    type="textarea"
    envVar="EXTRA_KEYS"
    bind:value={extraKeys}
    {rows}
    original={$profile.config.extraKeys?.join('\n') ?? ''} />
  <Input
    id="config.hostId"
    envVar="HOST_ID"
    bind:value={config.hostId}
    original={$profile.config.hostId} />

  <!-- Network Section -->
  <Network bind:config original={$profile.config} />
  <!-- System Section -->
  <System bind:config original={$profile.config} />
  <!-- SSL Section -->
  <SslConfig bind:config original={$profile.config} />
  <!-- Services Section -->
  <Services bind:config original={$profile.config} />
  <!-- Logging Section -->
  <Logging bind:config original={$profile.config} />
</CardBody>

<Footer {submit} saveDisabled={!nav.formChanged} />

<Nodal bind:isOpen title="Raw Configuration" size="xl" full>
  <pre>{JSON.stringify($profile.config, null, 2)}</pre>
</Nodal>
