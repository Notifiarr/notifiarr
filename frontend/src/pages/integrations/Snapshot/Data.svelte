<script lang="ts">
  import { Card, CardHeader, Table, Col } from '@sveltestrap/sveltestrap'
  import type { Snapshot } from '../../../api/notifiarrConfig'
  import { age, formatBytes } from '../../../includes/util'
  import { profile } from '../../../api/profile.svelte'
  import Modal from '../Modal.svelte'
  import T, { _ } from '../../../includes/Translate.svelte'

  type Props = { snapshot: Snapshot; snapshotAge: Date }
  let { snapshot, snapshotAge }: Props = $props()
  let snapshotModal: Modal | null = $state(null)
</script>

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
            <td class="text-break">Nvidia GPUs</td>
            <td class="text-break">{snapshot.nvidia?.length ?? 0}</td>
          </tr>
        {/if}
        {#if snapshot.mysql?.length}
          <tr>
            <td class="text-break">MySQL Servers</td>
            <td class="text-break">{Object.keys(snapshot.mysql ?? {}).length ?? 0}</td>
          </tr>
        {/if}
        {#if snapshot.ipmiSensors?.length}
          <tr>
            <td class="text-break">IPMI Sensors</td>
            <td class="text-break">{snapshot.ipmiSensors?.length ?? 0}</td>
          </tr>
        {/if}
        {#if snapshot.synology?.vender}
          <tr>
            <td class="text-break">Synology</td>
            <td class="text-break">{snapshot.synology?.vender}</td>
          </tr>
        {/if}
        <tr>
          <td class="text-break">Raid Configs</td>
          <td class="text-break">
            {snapshot.raid?.megacli?.length ?? 0 + (snapshot.raid?.mdstat ? 1 : 0)}</td>
        </tr>
        {#if snapshot.zfsPools?.length}
          <tr>
            <td class="text-break">ZFS Pools</td>
            <td class="text-break">{Object.keys(snapshot.zfsPools ?? {}).length ?? 0}</td>
          </tr>
        {/if}
        <tr>
          <td class="text-break"><go-to page="ProcessList">Processes</go-to></td>
          <td class="text-break">{snapshot.processes?.length ?? 0}</td>
        </tr>
        <tr>
          <td class="text-break">Drive Temps</td>
          <td class="text-break">
            {Object.keys(snapshot.driveTemps ?? {}).length ?? 0}</td>
        </tr>
        <tr>
          <td class="text-break">Drive Health</td>
          <td class="text-break">
            {Object.keys(snapshot.driveHealth ?? {}).length ?? 0}</td>
        </tr>
        <tr>
          <td class="text-break">Disk Usage</td>
          <td class="text-break">{Object.keys(snapshot.diskUsage ?? {}).length ?? 0}</td>
        </tr>
        {#if snapshot.quotas?.length}
          <tr>
            <td class="text-break">Disk Quotas</td>
            <td class="text-break">{Object.keys(snapshot.quotas ?? {}).length ?? 0}</td>
          </tr>
        {/if}
        <tr>
          <td class="text-break">IO Top</td>
          <td class="text-break">{snapshot.ioTop ? 'yes' : 'no'}</td>
        </tr>
        <tr>
          <td class="text-break">IO Stat</td>
          <td class="text-break">{snapshot.ioStat?.length ?? 0}</td>
        </tr>
        <tr>
          <td class="text-break">IO Stat 2</td>
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
