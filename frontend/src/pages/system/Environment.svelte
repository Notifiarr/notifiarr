<script lang="ts">
  import { faCubes } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { profile } from '../../api/profile.svelte'
  import { Table } from '@sveltestrap/sveltestrap'
  import Header from '../../includes/Helem.svelte'
</script>

<!-- Environment Section -->
<Header id="EnvironmentVariables" i={faCubes} c1="tan" c2="orange" />
<Table>
  <tbody>
    {#each Object.entries($profile.environment || {}) as [key, value]}
      <tr>
        <th>{key}</th>
        <td class="text-break">{value}</td>
      </tr>
    {/each}
  </tbody>
</Table>
