<script lang="ts">
  import { faCompactDisc } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from '../../includes/Helem.svelte'
  import { formatBytes } from '../../includes/util'
  import { _ } from '../../includes/Translate.svelte'

  let free = $derived($_('system.StorageData.Free'))
  let used = $derived($_('system.StorageData.Used'))
  let total = $derived($_('system.StorageData.Total'))
  let fs = $derived($_('system.StorageData.FS'))
</script>

<Header
  id="StorageData"
  i={faCompactDisc}
  c1="darkblue"
  c2="lightblue"
  d1="cyan"
  d2="lightcyan" />

<Table>
  <thead>
    <tr>
      <th>{$_('system.StorageData.Device')}</th>
      <th>{$_('system.StorageData.Mount')}</th>
      <th>{$_('words.select-option.Information')}</th>
    </tr>
  </thead>
  <tbody>
    {#each Object.entries($profile.disks || {}) as [mapKey, disk]}
      <tr>
        <th>{disk?.device ?? mapKey ?? '—'}</th>
        <th>{disk?.name ?? '-'}</th>
        <td>
          <div class="row">
            <div class="col-md-3">{total}: {formatBytes(disk?.total || 0)}</div>
            <div class="col-md-3">{free}: {formatBytes(disk?.free || 0)}</div>
            <div class="col-md-3">{used}: {formatBytes(disk?.used || 0)}</div>
            <div class="col-md-3">
              {fs}: {disk?.fsType}{#if disk?.opts?.length},
                {disk.opts.join(', ')}{/if}
            </div>
          </div>
        </td>
      </tr>
    {/each}
  </tbody>
</Table>
