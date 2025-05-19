<script lang="ts">
  import golift from '../../assets/golift.png?inline'
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from './Header.svelte'
  import T from '../../includes/Translate.svelte'
</script>

<!-- Build Section -->
<Header id="BuildInformation" logo={golift} />
<Table>
  <tbody>
    <tr>
      <th><T id="system.BuildInformation.Branch" /></th>
      <td>{$profile.branch}</td>
    </tr>
    <tr>
      <th><T id="system.BuildInformation.BuildUser" /></th>
      <td>{$profile.buildUser}</td>
    </tr>
    <tr>
      <th><T id="system.BuildInformation.BuildDate" /></th>
      <td>{new Date($profile.buildDate).toLocaleString()}</td>
    </tr>
    <tr>
      <th><T id="system.BuildInformation.GoVersion" /></th>
      <td>{$profile.goVersion}</td>
    </tr>
    <tr>
      <th><T id="system.BuildInformation.Integrity" /></th>
      <td>
        {$profile.clientInfo?.integrityCheck ? 'Passed' : 'Failed'} ({$profile.md5})
      </td>
    </tr>
  </tbody>
</Table>
