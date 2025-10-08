<script lang="ts">
  import {
    Button,
    Card,
    CardBody,
    CardFooter,
    CardHeader,
    InputGroup,
    InputGroupText,
    Input,
  } from '@sveltestrap/sveltestrap'
  import { getUi } from '../../api/fetch'
  import type { BrowseDir } from '../../api/notifiarrConfig'
  import { rtrim } from '../util'
  import Fa from '../Fa.svelte'
  import { faArrowUpToArc } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import T, { _ } from '../Translate.svelte'
  import { tick, type Snippet } from 'svelte'
  import FileList from './FileList.svelte'
  import { faCheck, faSpinner } from '@fortawesome/sharp-duotone-regular-svg-icons'

  type Props = {
    /**
     * This represents the selected file.
     * If a value is set, the path will be set to the parent directory of the value.
     */
    value: string
    /** This is called when the user clicks a close button. It should destroy this component. */
    close: () => void
    /** When set to true, disallows files from being selected, and won't even show them. */
    dir?: boolean
    /** When set to true, disallows directories from being selected. */
    file?: boolean
    /** The height of the card. */
    height?: string
    /** When set to true, shows a cancel button. */
    showCancel?: boolean
    /** Children to render in the card header. */
    children?: Snippet
  }

  let {
    value = $bindable(),
    close,
    dir = false,
    file = false,
    height = '100%',
    showCancel = false,
    children,
  }: Props = $props()

  let wd: BrowseDir = $state({ path: value, files: [], dirs: [], sep: '/', mom: '' })
  let selected = $state(value || '/')
  let respErr = $state('')
  let loading = $state(false)

  const cd = (e: Event, to: string, direct = false) => {
    e.preventDefault()
    selected = direct ? to : rtrim(wd.path, wd.sep) + wd.sep + to
  }

  const select = (e: Event, file: string, dir = false) => {
    e.preventDefault()
    value = (dir ? '' : rtrim(wd.path, wd.sep) + wd.sep) + file
    close()
  }

  const getFiles = async () => {
    loading = true
    const resp = await getUi('browse?dir=' + selected, true)
    if (resp.ok) {
      wd = resp.body as BrowseDir
      respErr = ''
    } else {
      // Set the path to the selected path, and clear the files and directories.
      // Makes navigating out of an error state easier.
      wd.mom = wd.path
      wd.path = selected
      wd.dirs = wd.files = undefined
      await tick()
      respErr = resp.body
    }
    loading = false
  }

  $effect(() => {
    if (selected) getFiles()
  })
</script>

<Card style="height: {height};min-height: 280px;">
  <CardHeader>
    <!-- Path title (input group). -->
    <form onsubmit={e => cd(e, wd.path, true)}>
      <InputGroup>
        <Button
          outline
          onclick={e => cd(e, wd.mom || wd.sep, true)}
          disabled={loading}
          type="button">
          {#if loading}
            <!-- Loading spinner. -->
            <Fa i={faSpinner} c1="steelblue" d2="firebrick" scale={1.5} spin />
          {:else}
            <!-- Up button. -->
            <Fa i={faArrowUpToArc} c1="steelblue" d2="firebrick" scale={1.5} />
          {/if}
        </Button>
        <InputGroupText><T id="LogFiles.titles.Path" /></InputGroupText>
        <Input bind:value={wd.path} />
        <!-- Go button. -->
        {#if wd.path !== selected}
          <Button type="submit" color="primary" outline><T id="FileBrowser.Go" /></Button>
        {/if}
        <!-- Select folder button. -->
        {#if !file}
          <Button
            title={$_('FileBrowser.SelectFolder', { values: { path: wd.path } })}
            color="success"
            outline
            type="button"
            onclick={e => select(e, wd.path, true)}>
            <Fa i={faCheck} c1="limegreen" d1="green" scale={1.5} />
          </Button>
        {/if}
      </InputGroup>
    </form>

    <!-- Children (a description of the file browser). -->
    {@render children?.()}
  </CardHeader>

  <CardBody class="overflow-auto h-100 p-0">
    <!-- Error message. -->
    {#if respErr}
      <Card outline color="danger" class="m-2 text-center" body>{respErr}</Card>
    {/if}
    <FileList {wd} {dir} {cd} {select} />
  </CardBody>

  <CardFooter class="clearfix">
    <ul class="d-inline-block mb-0 ps-2">
      <li><T id="FileBrowser.Folders" count={wd.dirs?.length ?? 0} /></li>
      <li><T id="FileBrowser.Files" count={wd.files?.length ?? 0} /></li>
      {#if value}<li><T id="FileBrowser.Selected" path={value} /></li>{/if}
    </ul>

    {#if showCancel}
      <!-- Cancel button. -->
      <Button color="secondary" outline onclick={close} class="float-end mx-2">
        <T id="buttons.Cancel" /></Button>
    {/if}
  </CardFooter>
</Card>
