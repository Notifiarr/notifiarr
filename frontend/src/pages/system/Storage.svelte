<script lang="ts">
  import { faCompactDisc } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from './Header.svelte'
  import { formatBytes } from '../../includes/util'
  import { _ } from '../../includes/Translate.svelte'
  import { get } from 'svelte/store'

  let free = $derived(get(_)('system.StorageData.Free'))
  let used = $derived(get(_)('system.StorageData.Used'))
  let total = $derived(get(_)('system.StorageData.Total'))
  let fs = $derived(get(_)('system.StorageData.FS'))
  let ro = $derived(get(_)('system.StorageData.ro'))
</script>

<Header
  id="StorageData"
  i={faCompactDisc}
  c1="darkblue"
  c2="lightblue"
  d1="cyan"
  d2="lightcyan" />

<Table>
  <tbody>
    {#each Object.entries($profile.disks || {}) as [device, disk]}
      <tr>
        <th>{device}</th>
        <td>
          <div class="row">
            <div class="col-md-3">{total}: {formatBytes(disk?.total || 0)}</div>
            <div class="col-md-3">{free}: {formatBytes(disk?.free || 0)}</div>
            <div class="col-md-3">{used}: {formatBytes(disk?.used || 0)}</div>
            <div class="col-md-3">
              {fs}: {disk?.fsType}
              {#if disk?.readOnly}
                <span class="text-muted">({ro})</span>
              {/if}
            </div>
          </div>
        </td>
      </tr>
    {/each}
  </tbody>
</Table>
