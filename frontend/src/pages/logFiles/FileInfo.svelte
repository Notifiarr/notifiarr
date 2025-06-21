<script lang="ts">
  import {
    Button,
    ButtonGroup,
    Card,
    Modal,
    ModalBody,
    ModalFooter,
    ModalHeader,
    Table,
  } from '@sveltestrap/sveltestrap'
  import type { LogFileInfo } from '../../api/notifiarrConfig'
  import { profile } from '../../api/profile.svelte'
  import T, { _, datetime } from '../../includes/Translate.svelte'
  import {
    faCloudDownload,
    faCloudUpload,
    faTrashAlt,
    faArrowUpFromBracket,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { faQuestionCircle } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import Fa from '../../includes/Fa.svelte'
  import { getUi } from '../../api/fetch'
  import { success, warning } from '../../includes/util'
  import { slide } from 'svelte/transition'
  import { theme } from '../../includes/theme.svelte'

  let { file, list = $bindable() }: { file: LogFileInfo; list: LogFileInfo[] } = $props()

  let showTooltip = $state(false)
  let showDelModal = $state(false)

  const deleteFile = async () => {
    if (file.used) return
    const resp = await getUi(`deleteFile/logs/${file.id}`, false)
    if (resp.ok) {
      success($_('LogFiles.deleteSuccess', { values: { file: file.path } }))
      list = list.filter(f => f.id !== file.id)
    } else {
      warning($_('LogFiles.deleteError', { values: { error: resp.body } }))
    }
  }

  const downloadFile = async () => {
    window.location.href = 'downloadFile/logs/' + file.id
  }

  const uploadFile = async () => {
    const resp = await getUi(`uploadFile/logs/${file.id}`, false)
    if (resp.ok) {
      success($_('LogFiles.uploadSuccess', { values: { file: file.path } }))
    } else {
      warning($_('LogFiles.uploadError', { values: { error: resp.body } }))
    }
  }
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
    <Button
      color="primary"
      size="sm"
      title={$_('LogFiles.button.download')}
      outline
      onclick={downloadFile}>
      <Fa i={faCloudDownload} scale={1.5} />&nbsp; <T id="LogFiles.button.download" />
    </Button>
    <Button
      color="primary"
      size="sm"
      title={$_('LogFiles.button.upload')}
      outline
      onclick={uploadFile}>
      <Fa i={faCloudUpload} scale={1.5} />&nbsp; <T id="LogFiles.button.upload" />
    </Button>
  </ButtonGroup>
</div>
<div class="mt-2 d-inline-block">
  <ButtonGroup>
    <Button
      color="secondary"
      size="sm"
      onclick={() => (showTooltip = !showTooltip)}
      outline
      title={$_('phrases.ShowMore')}>
      {#if showTooltip}
        <Fa i={faArrowUpFromBracket} c1="gray" d1="gainsboro" c2="orange" scale="1.5x" />
      {:else}
        <Fa i={faQuestionCircle} c1="gray" d1="gainsboro" c2="orange" scale="1.5x" />
      {/if}
    </Button>
    <Button
      color="danger"
      size="sm"
      title={$_('LogFiles.button.delete')}
      outline
      disabled={file.used}
      onclick={() => (showDelModal = true)}>
      <Fa i={faTrashAlt} scale={1.5} />&nbsp; <T id="LogFiles.button.delete" />
    </Button>
  </ButtonGroup>
</div>

{#if showTooltip}
  <div transition:slide>
    <Card body class="mt-1" color="warning" outline>
      <p class="mb-0"><T id="LogFiles.button.tooltip" /></p>
    </Card>
  </div>
{/if}

<Modal
  isOpen={showDelModal}
  toggle={() => (showDelModal = false)}
  theme={$theme}
  centered>
  <ModalHeader>
    <T id="LogFiles.confirmDelete" file={file.path} />
  </ModalHeader>
  <ModalBody><T id="LogFiles.deleteConfirm" /></ModalBody>
  <ModalFooter>
    <Button color="danger" outline onclick={deleteFile}>
      <T id="LogFiles.button.delete" />
    </Button>
    <Button color="primary" outline onclick={() => (showDelModal = false)}>
      <T id="buttons.Cancel" />
    </Button>
  </ModalFooter>
</Modal>

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
