<script lang="ts">
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import { faSlidersH } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Header from '../../includes/Helem.svelte'
  import T from '../../includes/Translate.svelte'
</script>

<!-- Startup Parameters -->
<Header
  id="StartupParams"
  i={faSlidersH}
  c1="darkseagreen"
  c2="lightseagreen"
  d1="darkcyan"
  d2="lightcyan" />
<Table>
  <tbody>
    <tr>
      <th><T id="system.StartupParams.BinaryPath" /></th>
      <td>{$profile.binary}</td>
    </tr>
    <tr>
      <th><T id="system.StartupParams.ConfigFile" /></th>
      <td>{$profile.flags?.configFile}</td>
    </tr>
    <tr>
      <th><T id="system.StartupParams.EnvironmentPrefix" /></th>
      <td>{$profile.flags?.envPrefix}</td>
    </tr>
    {#if $profile.flags?.extraConf?.length}
      <tr>
        <th><T id="system.StartupParams.ExtraConfigFiles" /></th>
        <td>{$profile.flags.extraConf.join(', ')}</td>
      </tr>
    {/if}
    <tr>
      <th><T id="system.StartupParams.UserID" /></th>
      <td>{$profile.uid}</td>
    </tr>
    <tr>
      <th><T id="system.StartupParams.GroupID" /></th>
      <td>{$profile.gid}</td>
    </tr>
    <tr>
      <th><T id="system.StartupParams.IPAddress" /></th>
      <td>{$profile.ip}</td>
    </tr>
  </tbody>
</Table>
