<script lang="ts">
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from '../../includes/Helem.svelte'
  import T from '../../includes/Translate.svelte'
  import {
    faDocker,
    faWindows,
    faLinux,
    faFreebsd,
    faApple,
    faNfcSymbol,
  } from '@fortawesome/free-brands-svg-icons'
  import type { Props } from '../../includes/Fa.svelte'

  let icon: Props = { i: faNfcSymbol }

  if ($profile.isDocker) {
    icon.i = faDocker
    icon.c1 = 'blue'
    icon.style = 'margin-bottom: 1px;'
  } else if ($profile.isWindows) {
    icon.i = faWindows
    icon.c1 = 'darkgoldenrod'
  } else if ($profile.isLinux) {
    icon.i = faLinux
    icon.c1 = 'lightcoral'
  } else if ($profile.isFreeBsd) {
    icon.i = faFreebsd
    icon.c1 = 'orange'
  } else if ($profile.isDarwin) {
    icon.i = faApple
    icon.c1 = 'orange'
  }
</script>

<!-- OS Section -->
<Header id="OperatingSystem" {...icon} />
<Table>
  <tbody>
    <tr>
      <th><T id="system.OperatingSystem.Hostname" /></th>
      <td>{$profile.hostInfo?.hostname}</td>
    </tr>
    <tr>
      <th><T id="system.OperatingSystem.UniqueID" /></th>
      <td>{$profile.hostInfo?.hostId}</td>
    </tr>
    <tr>
      <th><T id="system.OperatingSystem.Platform" /></th>
      <td>
        {$profile.os}
        {$profile.hostInfo?.platformVersion} ({$profile.arch})
      </td>
    </tr>
    {#if $profile.hostInfo?.kernelVersion}
      <tr>
        <th><T id="system.OperatingSystem.KernelVersion" /></th>
        <td>{$profile.hostInfo.kernelVersion}</td>
      </tr>
    {/if}
    {#if $profile.hostInfo?.virtualizationSystem}
      <tr>
        <th><T id="system.OperatingSystem.VirtualizationSystem" /></th>
        <td>{$profile.hostInfo.virtualizationSystem}</td>
      </tr>
    {/if}
  </tbody>
</Table>
