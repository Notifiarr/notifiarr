<script lang="ts">
  import { Table } from '@sveltestrap/sveltestrap'
  import Storage from '../../system/Storage.svelte'
  import { age } from '../../../includes/util'
  import H from '../../../includes/Helem.svelte'
  import { faHardDrive, faHeartPulse } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import T from '../../../includes/Translate.svelte'

  type Props = {
    driveTemps: Record<string, number>
    driveHealth: Record<string, string>
    driveAges: Record<string, number>
    mdstat: string[][]
    // megacli???
  }
  const { driveTemps, driveHealth, driveAges, mdstat }: Props = $props()

  const mdstatRecords = $derived(
    mdstat.reduce(
      (acc, line) => {
        if (line?.length > 1 && line[0]) acc[line[0]] = line[1]
        return acc
      },
      {} as Record<string, string>,
    ),
  )

  const numDrives = $derived(
    Math.max(
      parseInt(mdstatRecords['sbNumDisks'] ?? '0'),
      parseInt(mdstatRecords['mdNumDisks'] ?? '0'),
    ),
  )

  // Get all unique drive names from all three records
  const allDriveNames = $derived(
    new Set([
      ...Object.keys(driveTemps),
      ...Object.keys(driveHealth),
      ...Object.keys(driveAges),
    ]),
  )

  const allDrives = $derived(
    Array.from(allDriveNames).map(driveName => ({
      name: driveName,
      temp: driveTemps[driveName] ?? 0,
      health: driveHealth[driveName] ?? 'Unknown',
      age: driveAges[driveName] ?? 0,
    })),
  )
</script>

<!-- Drive Information -->
<H
  parent="Integrations.DriveData"
  id="driveInformation"
  i={faHeartPulse}
  c1="darkblue"
  c2="lightblue"
  d1="cyan"
  d2="lightcyan" />

<Table size="sm" striped borderless>
  <thead>
    <tr>
      <th><T id="Integrations.DriveData.titles.Drive" /></th>
      <th><T id="Integrations.DriveData.titles.Temp" /></th>
      <th><T id="Integrations.DriveData.titles.Health" /></th>
      <th><T id="Integrations.DriveData.titles.Age" /></th>
    </tr>
  </thead>
  <tbody>
    {#each allDrives as drive}
      <tr>
        <td>{drive.name}</td>
        <td>{drive.temp}°C</td>
        <td>{drive.health}</td>
        <td>{age(drive.age * 60 * 60 * 1000)}</td>
      </tr>
    {/each}
  </tbody>
</Table>

<!-- Disk Storage -->
<Storage />

<!-- MDSTAT -->
<H
  parent="Integrations.DriveData"
  id="mdStatRaid"
  i={faHardDrive}
  c1="darkblue"
  c2="lightblue"
  d1="cyan"
  d2="lightcyan" />

<Table size="sm" striped borderless>
  <thead>
    <tr>
      <th><T id="Integrations.DriveData.titles.Setting" /></th>
      <th><T id="Integrations.DriveData.titles.Value" /></th>
    </tr>
  </thead>
  <tbody>
    {#each Object.entries(mdstatRecords) as [setting, value]}
      <!-- Some settings end with .0, .1, etc to mark the drive number. -->
      {@const driveNum = parseInt(setting.match(/\.([0-9]+)$/)?.[1] ?? '-1')}
      {#if value && driveNum < numDrives}
        <tr><td>{setting}</td><td>{value}</td></tr>
      {/if}
    {/each}
  </tbody>
</Table>
