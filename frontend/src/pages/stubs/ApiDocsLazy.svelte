<!-- Thin registry shell. RapiDoc and the OpenAPI specs stay out of the boot bundle. -->
<script lang="ts" module>
  export const page = { id: 'ApiDocs' }
</script>

<script lang="ts">
  import { CardBody } from '@sveltestrap/sveltestrap'
  import Fa from '../../includes/Fa.svelte'
  import T from '../../includes/Translate.svelte'
  import Warning from 'phosphor-svelte/lib/Warning'
  import CircleNotch from 'phosphor-svelte/lib/CircleNotch'

  const loaded = import('./ApiDocs.svelte')
</script>

{#await loaded}
  <CardBody>
    <h5>
      <Fa i={CircleNotch} spin weight="bold" btn c1="orange">
        <T id="phrases.Loading" />
      </Fa>
    </h5>
  </CardBody>
{:then mod}
  <mod.default />
{:catch error}
  <CardBody>
    <h5>
      <Fa i={Warning} btn c1="red">
        <T id="phrases.ERROR" />
      </Fa><br />
      {error instanceof Error ? error.message : `${error}`}
    </h5>
  </CardBody>
{/await}
