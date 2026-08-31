<script lang="ts" module>
  import {
    faBook,
    faExclamationTriangle,
  } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'ApiDocs',
    i: faBook,
    c1: 'lightcoral',
    c2: 'lightblue',
    d1: 'lightblue',
    d2: 'lightcoral',
  }
</script>

<script lang="ts">
  import { onDestroy, onMount, tick } from 'svelte'
  import T from '../../includes/Translate.svelte'
  import { theme } from '../../includes/theme.svelte'
  import { urlbase } from '../../api/fetch'
  import { CardBody, Input } from '@sveltestrap/sveltestrap'
  import Fa from '../../includes/Fa.svelte'
  import { faSpinner } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { profile } from '../../api/profile.svelte'
  import Header from '../../includes/Header.svelte'

  const apiDocs = [
    { id: 'api', file: 'api_swagger.json', path: 'api' },
    { id: 'ui', file: 'ui_swagger.json', path: 'ui' },
  ]

  let loadError = $state('')
  let doc = $state(apiDocs[0])
  let ui = $state<any>()

  // Keep the last fetched spec. Swagger UI's internal state.json often drops
  // `swagger`/`openapi`; rewriting from that triggers the missing-version error.
  let currentSpec: Record<string, any> | undefined
  let loadSeq = 0
  let fetchAbort: AbortController | undefined
  let uiInit: Promise<any> | undefined

  // https://github.com/swagger-api/swagger-ui/issues/5981
  const UrlMutatorPlugin = (system: any) => ({
    rootInjects: {
      setBasePath: (basePath: string) => {
        if (doc.id === 'api')
          system.preauthorizeApiKey('ApiKeyAuth', $profile.config.apiKey)
        if (!currentSpec) return
        currentSpec = { ...currentSpec, basePath }
        return system.specActions.updateJsonSpec(currentSpec)
      },
    },
  })

  // Swagger UI 5 requires info.version. Stamp it from golift.io/version (profile).
  const specVersion = () =>
    [$profile.version, $profile.revision].filter(Boolean).join('-') || '0.0.0'

  const stampSpec = (spec: any) => ({
    ...spec,
    info: { ...spec.info, version: specVersion() },
  })

  const fetchSpec = async (
    selected: (typeof apiDocs)[number],
    signal: AbortSignal,
  ) => {
    const res = await fetch($urlbase + selected.file, { signal })
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
    return stampSpec(await res.json())
  }

  const showSpec = (
    spec: Record<string, any>,
    selected: (typeof apiDocs)[number],
  ) => {
    currentSpec = spec
    ui?.specActions.updateJsonSpec(spec)
    ui?.setBasePath($urlbase + selected.path)
  }

  const ensureUi = async (
    spec: Record<string, any>,
    selected: (typeof apiDocs)[number],
    seq: number,
  ) => {
    uiInit ??= (async () => {
      await import('swagger-ui/dist/swagger-ui.css')
      const SwaggerUI = await import('swagger-ui')
      return SwaggerUI.default({
        spec,
        plugins: [UrlMutatorPlugin],
        defaultModelsExpandDepth: 0,
        dom_id: '#swagger-ui-container',
        onComplete: () => {
          if (seq !== loadSeq || !ui) return
          ui.setBasePath($urlbase + selected.path)
        },
      })
    })()

    try {
      ui = await uiInit
    } catch (error) {
      uiInit = undefined
      throw error
    }
    if (seq !== loadSeq) return
    showSpec(spec, selected)
  }

  const loadSelected = async () => {
    const selected = doc
    const seq = ++loadSeq
    fetchAbort?.abort()
    fetchAbort = new AbortController()
    const { signal } = fetchAbort
    loadError = ''
    try {
      const spec = await fetchSpec(selected, signal)
      if (seq !== loadSeq) return
      await ensureUi(spec, selected, seq)
    } catch (error) {
      if (signal.aborted || seq !== loadSeq) return
      loadError = error instanceof Error ? error.message : `${error}`
    }
  }

  onDestroy(() => fetchAbort?.abort())

  onMount(async () => {
    await tick()
    await loadSelected()
  })
</script>

<Header {page} badge="v{$profile.version}">
  <p><T id="ApiDocs.Contrast" /></p>
  <Input type="select" bind:value={doc} onchange={loadSelected} class="mb-2">
    <option value={null} disabled><T id="ApiDocs.Choose" /></option>
    {#each apiDocs as ad}
      <option value={ad} selected={ad.id === doc.id}>
        <T id={`ApiDocs.${ad.id}.title`} basePath={$urlbase + ad.path} />
      </option>
    {/each}
  </Input>
  <T id={`ApiDocs.${doc.id}.body`} />
  <ul class="mb-0 mt-2">
    <li><T id="ApiDocs.BasePath" basePath={$urlbase + doc.path} /></li>
  </ul>
</Header>

<div id="swagger-ui-container" class:dark-mode={theme.isDark}>
  <CardBody>
    <h5>
      {#if loadError}
        <Fa i={faExclamationTriangle} class="me-2" scale={1.5} c1="red" />
        <T id="phrases.ERROR" /><br />
        {loadError}
      {:else}
        <Fa i={faSpinner} spin class="me-2" scale={1.5} c1="orange" />
        <T id="phrases.Loading" />
      {/if}
    </h5>
  </CardBody>
</div>

<style>
  #swagger-ui-container :global(.swagger-ui .info *),
  #swagger-ui-container :global(.opblock-tag),
  #swagger-ui-container :global(.opblock-summary-description),
  #swagger-ui-container :global(.opblock-description-wrapper *),
  #swagger-ui-container :global(.opblock-section-header *),
  #swagger-ui-container :global(.response-col_status),
  #swagger-ui-container :global(.responses-inner h4),
  #swagger-ui-container :global(.responses-inner h5),
  #swagger-ui-container :global(td),
  #swagger-ui-container :global(th),
  #swagger-ui-container :global(.model),
  #swagger-ui-container :global(.btn),
  #swagger-ui-container :global(.parameter__name),
  #swagger-ui-container :global(.parameter__type) {
    color: var(--bs-body-color) !important;
  }

  #swagger-ui-container :global(.wrapper) {
    max-width: none !important;
  }

  #swagger-ui-container :global(input),
  #swagger-ui-container :global(select),
  #swagger-ui-container :global(.content-type) {
    color: black !important;
  }

  #swagger-ui-container :global(.model-box-control),
  #swagger-ui-container :global(.models-control),
  #swagger-ui-container :global(.opblock-summary-control) {
    color: var(--bs-tertiary-color) !important;
  }

  #swagger-ui-container :global(.prop-type) {
    color: var(--bs-primary) !important;
  }

  #swagger-ui-container :global(.response-col_status .response-undocumented),
  #swagger-ui-container :global(.model-title) {
    color: var(--bs-secondary-color) !important;
  }

  #swagger-ui-container :global(.swagger-ui .info a),
  #swagger-ui-container :global(button.tablinks) {
    color: #2fa582 !important;
  }

  #swagger-ui-container :global(.swagger-ui .info a):hover,
  #swagger-ui-container :global(button.tablinks):hover {
    color: #3cd2a5 !important;
  }

  #swagger-ui-container :global(.dark-mode .swagger-ui .info a) {
    color: #3cd2a5 !important;
  }

  #swagger-ui-container :global(.dark-mode .swagger-ui .info a):hover {
    color: #2fa582 !important;
  }

  #swagger-ui-container :global(.scheme-container) {
    background: none !important;
  }

  #swagger-ui-container :global(.opblock-section-header) {
    margin: 0 !important;
    background: none !important;
  }

  #swagger-ui-container :global(h4) {
    border: none !important;
  }

  #swagger-ui-container :global(.information-container),
  #swagger-ui-container :global(.scheme-container) {
    display: none !important;
  }
</style>
