<script lang="ts">
  import golift from '../../assets/golift.png?inline'
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from './Helem.svelte'
  import T, { date } from '../../includes/Translate.svelte'
  import { age } from '../../includes/util'
  const buildDate = new Date($profile.buildDate).getTime()
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
      <td>{date($profile.buildDate)} ({age(profile.now - buildDate, false)})</td>
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
