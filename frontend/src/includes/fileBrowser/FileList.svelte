<script lang="ts">
  import type { FileBrowser } from './browser.svelte'
  type Props = { fb: FileBrowser; dir: boolean; filter?: string }
  const { fb, dir, filter }: Props = $props()
  const filt = $derived(filter?.toLowerCase() ?? '')
</script>

<!-- Directory list. -->
<ul class="p-0">
  {#each fb.wd.dirs?.filter(d => d.toLowerCase().includes(filt)) || [] as dir}
    <li class="px-2">
      <a href="#{fb.wd.path}{fb.wd.sep}{dir}" onclick={e => fb.cd(e, dir)}>{dir}</a><span
        class="text-muted">{fb.wd.sep}</span>
    </li>
  {/each}
  <!-- File list. -->
  {#if !dir}
    {#each fb.wd.files?.filter(f => f.toLowerCase().includes(filt)) || [] as file}
      <li class="px-2">
        <a href="#{fb.wd.path}{fb.wd.sep}{file}" onclick={e => fb.select(e, file)}>
          {file}</a>
      </li>
    {/each}
  {/if}
</ul>

<style>
  a {
    text-decoration: none;
  }
  ul {
    columns: auto 200px; /* Automatically create columns with a minimum width of 200px */
    column-gap: 0px; /* Space between columns */
    list-style: none; /* Remove default list styling */
  }
  ul li {
    break-inside: avoid-column;
    page-break-inside: avoid; /* For older browsers/print */
  }
  ul li:nth-child(even) {
    background-color: var(--bs-card-cap-bg); /* Alternate row background */
  }
  ul li:hover {
    background-color: var(--bs-tertiary-bg); /* Hover background */
  }
</style>
