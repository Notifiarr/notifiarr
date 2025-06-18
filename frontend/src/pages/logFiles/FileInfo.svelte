<script lang="ts">
  import { Button, ButtonGroup, Table } from '@sveltestrap/sveltestrap'
  import type { LogFileInfo } from '../../api/notifiarrConfig'
  import { profile } from '../../api/profile.svelte'
  import T, { _, datetime } from '../../includes/Translate.svelte'
  import {
    faCloudDownload,
    faCloudUpload,
    faTrashAlt,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from '../../includes/Fa.svelte'

  let { file }: { file: LogFileInfo } = $props()
</script>

<Table striped size="sm" class="mb-1">
  <tbody class="fit">
    <tr><th><T id="LogFiles.titles.Path" /></th><td><b>{file.path}</b></td></tr>
    {#if !$profile.isWindows}
      <tr><th><T id="LogFiles.titles.Mode" /></th><td>{file.mode}</td></tr>
      <tr><th><T id="LogFiles.titles.Owner" /></th><td>{file.user}</td></tr>
    {/if}
    <tr><th><T id="LogFiles.titles.Bytes" /></th><td>{file.size}</td></tr>
    <tr><th><T id="LogFiles.titles.Date" /></th><td>{datetime(file.time)}</td></tr>
    <tr><th><T id="LogFiles.titles.InUse" /></th><td>{file.used}</td></tr>
  </tbody>
</Table>

<div class="mt-2 d-inline-block">
  <ButtonGroup>
    <Button color="primary" size="sm" title={$_('LogFiles.button.download')} outline>
      <Fa i={faCloudDownload} scale={1.5} />&nbsp; <T id="LogFiles.button.download" />
    </Button>
    <Button color="primary" size="sm" title={$_('LogFiles.button.upload')} outline>
      <Fa i={faCloudUpload} scale={1.5} />&nbsp; <T id="LogFiles.button.upload" />
    </Button>
  </ButtonGroup>
</div>
<div class="mt-2 d-inline-block">
  <Button color="danger" size="sm" title={$_('LogFiles.button.delete')} outline>
    <Fa i={faTrashAlt} scale={1.5} />&nbsp; <T id="LogFiles.button.delete" />
  </Button>
</div>

<style>
  .fit th {
    width: 1%;
    white-space: nowrap;
  }
  .fit td {
    word-break: break-all;
    word-wrap: break-word;
  }
</style>
