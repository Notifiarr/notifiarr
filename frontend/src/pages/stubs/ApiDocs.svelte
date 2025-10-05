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
  import { onMount, tick } from 'svelte'
  import T from '../../includes/Translate.svelte'
  import { theme } from '../../includes/theme.svelte'
  import { urlbase } from '../../api/fetch'
  import { CardBody, Input } from '@sveltestrap/sveltestrap'
  import Fa from '../../includes/Fa.svelte'
  import { faSpinner } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { profile } from '../../api/profile.svelte'
  import Header from '../../includes/Header.svelte'

  const apiDocs = [
    { what: 'PrivateAPI', file: 'api_swagger.json', path: 'api' },
    { what: 'WebUI', file: 'ui_swagger.json', path: 'ui' },
  ]

  let loadError = $state('')
  let doc = $state(apiDocs[0])
  let ui: any

  // https://github.com/swagger-api/swagger-ui/issues/5981
  const UrlMutatorPlugin = (system: any) => ({
    rootInjects: {
      setBasePath: (basePath: string) => {
        const jsonSpec = system.getState().toJSON().spec.json
        return system.specActions.updateJsonSpec({ ...jsonSpec, basePath })
      },
    },
  })

  const onchange = () => {
    ui.specActions.updateUrl($urlbase + doc.file)
    ui.specActions.download($urlbase + doc.file)
    ui.setBasePath($urlbase + doc.path)
    if (doc.what === 'PrivateAPI')
      ui.preauthorizeApiKey('ApiKeyAuth', $profile.config.apiKey)
  }

  onMount(async () => {
    await tick()
    try {
      await import('swagger-ui/dist/swagger-ui.css')
      const SwaggerUI = await import('swagger-ui')

      ui = await SwaggerUI.default({
        url: $urlbase + doc.file,
        plugins: [UrlMutatorPlugin],
        defaultModelsExpandDepth: 0,
        dom_id: '#swagger-ui-container',
        onComplete: () => {
          ui.setBasePath($urlbase + doc.path)
          if (doc.what === 'PrivateAPI')
            ui.preauthorizeApiKey('ApiKeyAuth', $profile.config.apiKey)
        },
      })
    } catch (error) {
      loadError = error instanceof Error ? error.message : `${error}`
    }
  })
</script>

<Header {page} badge="v{$profile.version}">
  <p><T id="ApiDocs.Contrast" /></p>
  <Input type="select" bind:value={doc} {onchange} class="mb-2">
    <option value={null} disabled><T id="ApiDocs.Choose" /></option>
    {#each apiDocs as ad}
      <option value={ad} selected={ad.what === doc.what}>
        <T id={`ApiDocs.${ad.what}.title`} basePath={$urlbase + ad.path} />
      </option>
    {/each}
  </Input>
  <T id={`ApiDocs.${doc.what}.body`} />
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
