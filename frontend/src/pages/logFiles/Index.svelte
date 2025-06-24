<script lang="ts" module>
  import {
    faArrowsSpin,
    faPrintMagnifyingGlass,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  export const page = {
    id: 'LogFiles',
    i: faPrintMagnifyingGlass,
    c1: 'midnightblue',
    c2: 'darkslategray',
    d1: 'lightgreen',
    d2: 'gold',
  }
</script>

<script lang="ts">
  import { CardBody, Col, Popover, Row, Table } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import { formatBytes, since } from '../../includes/util'
  import FileInfo from './FileInfo.svelte'
  import type { LogFileInfo } from '../../api/notifiarrConfig'
  import Fa from '../../includes/Fa.svelte'
  import { faListTimeline } from '@fortawesome/sharp-duotone-light-svg-icons'
  import Content from './Content.svelte'
  import { theme } from '../../includes/theme.svelte'

  let activeFile: LogFileInfo | null = $state(null)
  let activeTail: LogFileInfo | null = $state(null)

  const viewFile = (e: MouseEvent, file: LogFileInfo) => {
    e.preventDefault()
    activeFile = file
    activeTail = null
  }

  const tailFile = (e: MouseEvent, file: LogFileInfo) => {
    e.stopPropagation()
    activeTail = file
    activeFile = null
  }

  $effect(() => {
    // This deactivates a file that gets deleted.
    if (!$profile.logFileInfo?.list?.find(f => f.id == activeFile?.id)) {
      activeFile = null
    }
  })
</script>

<Header {page} />
<CardBody>
  <Row>
    <Col md={7}>
      <h4><T id="LogFiles.titles.FileList" /></h4>
      <div style="max-height: 305px; overflow-y: auto">
        <Table size="sm" striped borderless hover>
          <thead>
            <tr>
              <Popover
                target="followTooltip"
                trigger="click"
                theme={$theme}
                placement="right">
                <T id="LogFiles.FollowTooltip" />
              </Popover>
              <th id="followTooltip">
                <a href="#moreInfo" onclick={e => e.preventDefault()}>
                  <Fa i={faArrowsSpin} scale={1.2} /></a>
              </th>
              <th>
                <T id="LogFiles.titles.Name" />
                <small class="text-muted">
                  <T
                    id="LogFiles.FilesInDirs"
                    fileCount={$profile.logFileInfo?.list?.length}
                    dirCount={$profile.logFileInfo?.dirs?.length} />
                </small>
              </th>
              <th class="text-nowrap">{formatBytes($profile.logFileInfo?.size ?? 0)}</th>
              <th><T id="LogFiles.titles.Age" /></th>
            </tr>
          </thead>

          <tbody>
            {#each $profile.logFileInfo?.list ?? [] as file}
              {@const isActive = activeFile?.id === file.id || activeTail?.id === file.id}
              {@const vals = { values: { fileName: file.name } }}
              <tr
                class="cursor-pointer"
                onclick={e => viewFile(e, file)}
                aria-label={$_('LogFiles.titles.OpenFile', vals)}
                title={$_('LogFiles.titles.OpenFile', vals)}>
                <th
                  class="fit {isActive ? 'isActive' : ''}"
                  aria-label={$_('LogFiles.titles.TailFile', vals)}
                  title={$_('LogFiles.titles.TailFile', vals)}
                  onclick={e => (file.used ? tailFile(e, file) : null)}>
                  {#if file.used}<Fa i={faListTimeline} scale={1.2} />{:else}&nbsp;{/if}
                </th>
                <td class:isActive>
                  <a
                    aria-label={$_('LogFiles.titles.OpenFile', vals)}
                    title={$_('LogFiles.titles.OpenFile', vals)}
                    href={file.path}
                    onclick={e => viewFile(e, file)}
                    class="text-decoration-none">
                    {file.name}
                  </a>
                </td>
                <td class="fit {isActive ? 'isActive' : ''}">
                  {formatBytes(file.size)}
                </td>
                <td class="fit {isActive ? 'isActive' : ''}">
                  {since(file.time).split(' ').slice(0, 2).join(' ')}
                </td>
              </tr>
            {/each}
          </tbody>
        </Table>
      </div>
    </Col>

    <Col md={5}>
      <h4><T id="LogFiles.titles.Details" /></h4>
      {#if (activeFile || activeTail) && $profile.logFileInfo && $profile.logFileInfo.list}
        <FileInfo
          file={activeFile || activeTail!}
          bind:list={$profile.logFileInfo.list} />
      {:else}
        <p class="text-muted"><T id="LogFiles.NoFileSelected" /></p>
      {/if}
    </Col>
  </Row>

  <Row>
    <Col md={12}>
      <h4>
        {#if activeTail}
          <T id="LogFiles.titles.FollowingFile" />
        {:else}
          <T id="LogFiles.titles.FileContent" />
        {/if}
      </h4>
      {#if activeFile}
        <Content file={activeFile} />
      {:else if activeTail}
        <Content file={activeTail} tail />
      {:else}
        <p class="text-muted"><T id="LogFiles.SelectFile" /></p>
      {/if}
    </Col>
  </Row>
</CardBody>

<style>
  .fit {
    width: 1%;
    white-space: nowrap;
  }

  .isActive {
    background-color: var(--bs-secondary-bg) !important;
    color: var(--bs-secondary-color) !important;
  }

  .cursor-pointer {
    cursor: pointer;
  }
</style>
