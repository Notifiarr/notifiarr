<script lang="ts">
  import { CardHeader, Badge } from '@sveltestrap/sveltestrap'
  import Fa, { type Props as FaProps } from './Fa.svelte'
  import { _ } from './Translate.svelte'
  import { get } from 'svelte/store'
  type Props = { badge?: string; page: FaProps }
  let { badge = undefined, page }: Props = $props()
  const description: string = $derived(get(_)('navigation.pageDescription.' + page.id))
</script>

<CardHeader>
  <h2 class="page-title">
    {$_('navigation.titles.' + page.id)}
    {#if badge}<Badge color="notifiarr">{badge}</Badge>{/if}
    <Fa {...page} />
  </h2>
  {#if description != 'navigation.pageDescription.' + page.id}
    {@html description}
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
