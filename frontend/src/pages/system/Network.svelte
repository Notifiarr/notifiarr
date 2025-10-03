<script lang="ts">
  import { faChartNetwork } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from '../../includes/Helem.svelte'
  import T from '../../includes/Translate.svelte'
</script>

<!-- Network Section -->
<Header
  id="NetworkInformation"
  i={faChartNetwork}
  c1="firebrick"
  c2="firebrick"
  d1="pink" />
<Table>
  <tbody>
    <tr>
      <th><T id="system.NetworkInformation.IPAddress" /></th>
      <td>{$profile.ip}</td>
    </tr>
    <tr>
      <th><T id="system.NetworkInformation.Gateway" /></th>
      <td>{$profile.gateway}</td>
    </tr>
    <tr>
      <th><T id="system.NetworkInformation.Netmask" /></th>
      <td>{$profile.netmask}</td>
    </tr>
    <tr>
      <th><T id="system.NetworkInformation.Interface" /></th>
      <td>{$profile.ifName}</td>
    </tr>
  </tbody>
</Table>
