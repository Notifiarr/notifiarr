<script lang="ts">
  import { Card, CardHeader, Table, Col } from '@sveltestrap/sveltestrap'
  import type { Snapshot } from '../../../api/notifiarrConfig'
  import { age } from '../../../includes/util'
  import { profile } from '../../../api/profile.svelte'
  import Modal from '../Modal.svelte'
  import T, { _ } from '../../../includes/Translate.svelte'
  import Drives from './Drives.svelte'

  type Props = { snapshot: Snapshot; snapshotAge: Date }
  let { snapshot, snapshotAge }: Props = $props()
  let snapshotModal: Modal | null = $state(null)
  let driveModal: Modal | null = $state(null)
</script>

<Modal pageId="DriveData" bind:this={driveModal}>
  <Drives
    driveTemps={snapshot.driveTemps ?? {}}
    driveHealth={snapshot.driveHealth ?? {}}
    driveAges={snapshot.driveAges ?? {}}
    mdstat={snapshot.raid?.mdstat?.split('\n').map(line => line.split('=', 2)) ?? []} />
</Modal>

<Col class="mb-2" sm={12} md={6}>
  <Card outline color="tertiary">
    <CardHeader>
      <h5 class="m-0"><T id="Integrations.Snapshot.Data" /></h5>
    </CardHeader>
    <Table class="rounded-bottom mb-0" size="sm">
      <tbody class="table-body">
        <tr>
          <td class="text-break">
            <Modal pageId="Snapshot" data={snapshot} bind:this={snapshotModal} />
            <a href="#snapshot" onclick={snapshotModal?.toggle}>
              <T id="Integrations.Snapshot.Age" /></a>
          </td>
          <td class="text-break" style="min-width: 30px;">
            {age(profile.now - new Date(snapshotAge).getTime())}</td>
        </tr>
        {#if snapshot.nvidia?.length}
          <tr>
            <td class="text-break"><T id="Integrations.Snapshot.titles.NvidiaGPUs" /></td>
            <td class="text-break">{snapshot.nvidia?.length ?? 0}</td>
          </tr>
        {/if}
        {#if snapshot.mysql?.length}
          <tr>
            <td class="text-break"
              ><T id="Integrations.Snapshot.titles.MySQLServers" /></td>
            <td class="text-break">{Object.keys(snapshot.mysql ?? {}).length ?? 0}</td>
          </tr>
        {/if}
        {#if snapshot.ipmiSensors?.length}
          <tr>
            <td class="text-break">
              <T id="Integrations.Snapshot.titles.IPMISensors" /></td>
            <td class="text-break">{snapshot.ipmiSensors?.length ?? 0}</td>
          </tr>
        {/if}
        {#if snapshot.synology?.vender}
          <tr>
            <td class="text-break"><T id="Integrations.Snapshot.titles.Synology" /></td>
            <td class="text-break">{snapshot.synology?.vender}</td>
          </tr>
        {/if}
        <tr>
          <td class="text-break">
            <a href="#driveData" onclick={driveModal?.toggle}>
              <T id="Integrations.Snapshot.titles.RaidConfigs" /></a>
          </td>
          <td class="text-break">
            {snapshot.raid?.megacli?.length ?? 0 + (snapshot.raid?.mdstat ? 1 : 0)}</td>
        </tr>
        {#if Object.keys(snapshot.zfsPools ?? {}).length}
          <tr>
            <td class="text-break">
              <a href="#driveData" onclick={driveModal?.toggle}>
                <T id="Integrations.Snapshot.titles.ZFSPools" /></a>
            </td>
            <td class="text-break">{Object.keys(snapshot.zfsPools ?? {}).length ?? 0}</td>
          </tr>
        {/if}
        <tr>
          <td class="text-break">
            <go-to page="ProcessList">
              <T id="Integrations.Snapshot.titles.Processes" /></go-to>
          </td>
          <td class="text-break">{snapshot.processes?.length ?? 0}</td>
        </tr>
        <tr>
          <td class="text-break">
            <a href="#driveData" onclick={driveModal?.toggle}>
              <T id="Integrations.Snapshot.titles.DriveTemps" /></a>
          </td>
          <td class="text-break">
            {Object.keys(snapshot.driveTemps ?? {}).length ?? 0}</td>
        </tr>
        <tr>
          <td class="text-break">
            <a href="#driveData" onclick={driveModal?.toggle}>
              <T id="Integrations.Snapshot.titles.DriveHealth" /></a>
          </td>
          <td class="text-break">
            {Object.keys(snapshot.driveHealth ?? {}).length ?? 0}</td>
        </tr>
        <tr>
          <td class="text-break">
            <a href="#driveData" onclick={driveModal?.toggle}>
              <T id="Integrations.Snapshot.titles.DiskUsage" /></a>
          </td>
          <td class="text-break">{Object.keys(snapshot.diskUsage ?? {}).length ?? 0}</td>
        </tr>
        {#if snapshot.quotas?.length}
          <tr>
            <td class="text-break"><T id="Integrations.Snapshot.titles.DiskQuotas" /></td>
            <td class="text-break">{Object.keys(snapshot.quotas ?? {}).length ?? 0}</td>
          </tr>
        {/if}
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.IOTop" /></td>
          <td class="text-break">{snapshot.ioTop ? 'yes' : 'no'}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.IOStat" /></td>
          <td class="text-break">{snapshot.ioStat?.length ?? 0}</td>
        </tr>
        <tr>
          <td class="text-break"><T id="Integrations.Snapshot.titles.IOStat2" /></td>
          <td class="text-break">{Object.keys(snapshot.ioStat2 ?? {}).length ?? 0}</td>
        </tr>
      </tbody>
    </Table>
  </Card>
</Col>

<style>
  .table-body :global(tr:last-of-type td) {
    border-bottom: 0 !important;
  }
</style>
