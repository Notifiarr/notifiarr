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
  import { CardBody, Col, Row, Table } from '@sveltestrap/sveltestrap'
  import T from '../../includes/Translate.svelte'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import { formatBytes, since } from '../../includes/util'
  import FileInfo from './FileInfo.svelte'
  import type { LogFileInfo } from '../../api/notifiarrConfig'
  import Fa from '../../includes/Fa.svelte'
  import { faListTimeline } from '@fortawesome/sharp-duotone-light-svg-icons'
  import Content from './Content.svelte'

  let activeFile: LogFileInfo | null = $state(null)

  const set = (event: MouseEvent, file: LogFileInfo) => {
    event.preventDefault()
    activeFile = file
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
      <div style="max-height: 300px; overflow-y: auto">
        <Table size="sm" striped>
          <thead>
            <tr>
              <th><Fa i={faArrowsSpin} /></th>
              <th>
                <T id="LogFiles.titles.Name" />
                <small class="text-muted">
                  <T
                    id="LogFiles.FilesInDirs"
                    files={$profile.logFileInfo?.list?.length}
                    dirs={$profile.logFileInfo?.dirs?.length} />
                </small>
              </th>
              <th class="text-nowrap">{formatBytes($profile.logFileInfo?.size ?? 0)}</th>
              <th><T id="LogFiles.titles.Age" /></th>
            </tr>
          </thead>

          <tbody>
            {#each $profile.logFileInfo?.list ?? [] as file}
              {@const isActive = activeFile?.id === file.id}
              <tr class="cursor-pointer" onclick={e => set(e, file)}>
                <th class="fit {isActive ? 'isActive' : ''}">
                  <Fa i={faListTimeline} />
                </th>
                <td class:isActive>
                  <a
                    href={file.path}
                    onclick={e => set(e, file)}
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
      {#if activeFile && $profile.logFileInfo && $profile.logFileInfo.list}
        <FileInfo file={activeFile} bind:list={$profile.logFileInfo.list} />
      {:else}
        <p class="text-muted"><T id="LogFiles.NoFileSelected" /></p>
      {/if}
    </Col>
  </Row>

  <Row>
    <Col md={12}>
      <h4><T id="LogFiles.titles.FileContent" /></h4>
      {#if activeFile}
        <Content file={activeFile} />
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

  .cursor-pointer:hover td,
  .cursor-pointer:hover th {
    background-color: var(--bs-secondary-bg) !important;
    color: var(--bs-secondary-color) !important;
  }
</style>
