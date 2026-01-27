<script lang="ts">
  import { CardHeader, Badge } from '@sveltestrap/sveltestrap'
  import Fa, { type Props as FaProps } from './Fa.svelte'
  import { _ } from './Translate.svelte'
  import { get } from 'svelte/store'
  import type { Snippet } from 'svelte'
  type Props = {
    badge?: string
    page: FaProps
    children?: any
    description?: string | Snippet
  }
  let {
    badge = undefined,
    page,
    children,
    description = get(_)('navigation.pageDescription.' + page.id),
  }: Props = $props()
</script>

<CardHeader>
  <h2 class="page-title">
    {$_('navigation.titles.' + page.id)}
    {#if badge}<Badge color="notifiarr">{badge}</Badge>{/if}
    <Fa {...page} />
  </h2>
  {#if typeof description === 'string' && description != 'navigation.pageDescription.' + page.id}
    {@html description}
  {:else if typeof description !== 'string' && description !== undefined}
    {@render description()}
  {/if}
  {#if children}
    <hr />
    {@render children()}
  {/if}
</CardHeader>

<style>
  /* Small badge positioned to top. */
  .page-title :global(.badge) {
    font-size: 9px;
    vertical-align: top;
  }

  /* Move the icons on page titles to the right. */
  .page-title :global(svg) {
    position: absolute;
    right: 1rem;
    top: 0.5rem;
    height: 2.5rem;
    width: 2.5rem;
    margin-bottom: 0;
  }
</style>
