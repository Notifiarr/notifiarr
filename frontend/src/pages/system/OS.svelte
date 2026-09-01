<script lang="ts">
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from '../../includes/Helem.svelte'
  import T from '../../includes/Translate.svelte'
  import apple from '../../assets/logos/apple.png'
  import docker from '../../assets/logos/docker.png'
  import freebsd from '../../assets/logos/freebsd.png'
  import linux from '../../assets/logos/linux.png'
  import windows from '../../assets/logos/windows.png'

  const logo = $derived(
    $profile.isDocker
      ? docker
      : $profile.isWindows
        ? windows
        : $profile.isLinux
          ? linux
          : $profile.isFreeBsd
            ? freebsd
            : $profile.isDarwin
              ? apple
              : undefined,
  )
</script>

<!-- OS Section -->
<Header id="OperatingSystem" {logo} />
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
