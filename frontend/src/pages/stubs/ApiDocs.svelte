<script lang="ts" module>
  import { faBook } from '@fortawesome/sharp-duotone-light-svg-icons'
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
  import { onMount } from 'svelte'
  import 'swagger-ui/dist/swagger-ui.css'
  import T from '../../includes/Translate.svelte'
  import { theme } from '../../includes/theme.svelte'
  import 'swagger-ui/dist/swagger-ui.css'
  import { urlbase } from '../../api/fetch'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import Fa from '../../includes/Fa.svelte'
  import { faSpinner } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { profile } from '../../api/profile.svelte'

  onMount(async () => {
    const SwaggerUI = await import('swagger-ui')
    const swaggerJson = await import('../../../public/api_swagger.json')
    await SwaggerUI.default({
      spec: { ...swaggerJson, basePath: $urlbase },
      dom_id: '#swagger-ui-container',
    }).preauthorizeApiKey('ApiKeyAuth', $profile.config.apiKey)
  })
</script>

<div id="swagger-ui-container" class:dark-mode={theme.isDark}>
  <CardBody>
    <h5>
      <Fa i={faSpinner} spin class="me-2" scale={1.5} c1="orange" c2="orange" />
      <T id="phrases.Loading" />
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
</style>
