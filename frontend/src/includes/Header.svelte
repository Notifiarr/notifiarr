<script lang="ts">
  import { CardHeader, Badge } from '@sveltestrap/sveltestrap'
  import Fa from './Fa.svelte'
  import Code from 'phosphor-svelte/lib/Code'
  import { _ } from './Translate.svelte'
  import type { Snippet } from 'svelte'
  type Props = {
    badge?: string
    page: { id: string }
    children?: any
    description?: string | Snippet
    onclick?: () => void
  }
  let { badge = undefined, page, children, description, onclick }: Props = $props()

  const key = $derived('navigation.pageDescription.' + page.id)
  const fallback = $derived($_(key))
</script>

<CardHeader>
  <h2 class="page-title">
    {$_('navigation.titles.' + page.id)}
    {#if badge}<Badge color="notifiarr">{badge}</Badge>{/if}
    {#if onclick}
      <a
        href="#rawConfig"
        title="Raw Configuration"
        onclick={e => (e.preventDefault(), onclick())}>
        <Fa i={Code} c1="slategray" d1="gainsboro" btn />
      </a>
    {/if}
  </h2>
  {#if typeof description === 'string' && description != key}
    {@html description}
  {:else if description === undefined && fallback != key}
    {@html fallback}
  {:else if typeof description !== 'string' && description !== undefined}
    {@render description()}
  {/if}
  {#if children}
    <hr />
    {@render children()}
  {/if}
</CardHeader>

<style>
  .page-title {
    position: relative;
  }

  /* Small badge positioned to top. */
  .page-title :global(.badge) {
    font-size: 9px;
    vertical-align: top;
  }

  /* Configuration raw-config control sits in the old identity-icon slot. */
  .page-title :global(a:has(.nr-icon)) {
    position: absolute;
    right: 1rem;
    top: 0.5rem;
  }
</style>
