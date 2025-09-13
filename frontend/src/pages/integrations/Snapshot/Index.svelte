<script lang="ts">
  import { Card, CardHeader, Table, Col } from '@sveltestrap/sveltestrap'
  import type { Snapshot } from '../../../api/notifiarrConfig'
  import { age, formatBytes } from '../../../includes/util'
  import T, { _ } from '../../../includes/Translate.svelte'
  import Data from './Data.svelte'

  type Props = { snapshot: Snapshot; snapshotAge: Date }
  let { snapshot, snapshotAge }: Props = $props()
</script>

<Col class="mb-2" sm={12} md={6}>
  <Card outline color="secondary">
    <CardHeader>
      <h5 class="m-0"><T id="Integrations.Snapshot.System" /></h5>
    </CardHeader>
    <Table class="rounded-bottom mb-0" size="sm">
      <tbody class="table-body">
        <tr>
          <td class="text-break"><T id="system.OperatingSystem.Hostname" /></td>
          <td class="text-break">{snapshot.system.hostname} ({snapshot.system.os})</td>
        </tr>
        <tr>
          <td class="text-break"><T id="system.OperatingSystem.Platform" /></td>
          <td class="text-break">
            {snapshot.system.platform} {snapshot.system.platformVersion}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="profile.username.label" /></td>
          <td class="text-break">{snapshot.system.username}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.CPUPercent" /></td>
          <td class="text-break">{snapshot.system.cpuPerc.toFixed(2)}%</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.MemoryFree" /></td>
          <td class="text-break">{formatBytes(snapshot.system.memFree)}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.MemoryUsed" /></td>
          <td class="text-break">{formatBytes(snapshot.system.memUsed)}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.MemoryTotal" /></td>
          <td class="text-break">{formatBytes(snapshot.system.memTotal)}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.UserCPUTime" /></td>
          <td class="text-break">{age(snapshot.system.cpuTime.user * 1000)}</td>
        </tr>
        <tr>
          <td class="text-break">
            <T id="Integrations.Snapshot.titles.SystemCPUTime" /></td>
          <td class="text-break">{age(snapshot.system.cpuTime.system * 1000)}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.IdleCPUTime" /></td>
          <td class="text-break">{age(snapshot.system.cpuTime.idle * 1000)}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.mediaTitles.Users" /></td>
          <td class="text-break">{snapshot.system.users}</td>
        </tr>
      </tbody>
    </Table>
  </Card>
</Col>

<Data {snapshot} {snapshotAge} />

<style>
  .table-body :global(tr:last-of-type td) {
    border-bottom: 0 !important;
  }
</style>
