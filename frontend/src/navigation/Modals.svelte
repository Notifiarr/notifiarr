<script lang="ts" module>
  import { others } from './pages'
  import type { SvelteComponent } from 'svelte'
  export const modal = $state<Record<string, SvelteComponent>>({})
</script>

{#each others.filter(p => p.type === 'modal') as page}
  <page.component bind:this={modal[page.id]} />
{/each}
