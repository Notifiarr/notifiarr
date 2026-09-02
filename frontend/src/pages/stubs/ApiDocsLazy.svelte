<!-- Thin registry shell. RapiDoc stays out of the boot bundle and off the page slide. -->
<script lang="ts" module>
  export const page = { id: 'ApiDocs' }
</script>

<script lang="ts">
  import { onMount, tick } from 'svelte'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import Fa from '../../includes/Fa.svelte'
  import T from '../../includes/Translate.svelte'
  import Header from '../../includes/Header.svelte'
  import Warning from 'phosphor-svelte/lib/Warning'
  import CircleNotch from 'phosphor-svelte/lib/CircleNotch'
  import { profile } from '../../api/profile.svelte'
  import { afterPageSlide } from '../../navigation/pageSlide'

  // Fetch the chunk during the slide; do not mount it until the intro ends.
  const prefetch = import('./ApiDocs.svelte')
  let reveal = $state(false)

  onMount(() => {
    let gone = false
    const slide = afterPageSlide()
    void (async () => {
      await tick()
      await slide
      if (!gone) reveal = true
    })()
    return () => {
      gone = true
    }
  })
</script>

{#snippet loading(error?: unknown)}
  <Header {page} badge="v{$profile.version}" />
  <CardBody>
    <h5>
      {#if error}
        <Fa i={Warning} btn c1="red">
          <T id="phrases.ERROR" />
        </Fa><br />
        {error instanceof Error ? error.message : `${error}`}
      {:else}
        <Fa i={CircleNotch} spin weight="bold" btn c1="orange">
          <T id="phrases.Loading" />
        </Fa>
      {/if}
    </h5>
  </CardBody>
{/snippet}

{#if reveal}
  {#await prefetch}
    {@render loading()}
  {:then mod}
    <mod.default />
  {:catch error}
    {@render loading(error)}
  {/await}
{:else}
  {@render loading()}
{/if}
