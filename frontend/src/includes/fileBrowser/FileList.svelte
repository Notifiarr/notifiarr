<script lang="ts">
  import type { BrowseDir } from '../../api/notifiarrConfig'

  type Pick = (e: MouseEvent, dir: string) => void
  type Props = { wd: BrowseDir; dir: boolean; cd: Pick; select: Pick }
  const { wd, dir, cd, select }: Props = $props()
</script>

<!-- Directory list. -->
<ul class="p-0">
  {#each wd.dirs || [] as dir}
    <li class="px-2">
      <a href="#{wd.path}{wd.sep}{dir}" onclick={e => cd(e, dir)}>{dir}</a><span
        class="text-muted">{wd.sep}</span>
    </li>
  {/each}
  <!-- File list. -->
  {#if !dir}
    {#each wd.files || [] as file}
      <li class="px-2">
        <a href="#{wd.path}{wd.sep}{file}" onclick={e => select(e, file)}> {file}</a>
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
