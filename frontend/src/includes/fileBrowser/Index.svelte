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
    Tooltip,
  } from '@sveltestrap/sveltestrap'
  import Fa from '../Fa.svelte'
  import { faArrowUpToArc } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import T from '../Translate.svelte'
  import { type Snippet } from 'svelte'
  import FileList from './FileList.svelte'
  import { faCheck, faSpinner } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { FileBrowser } from './browser.svelte'
  import ActionBar from './ActionBar.svelte'

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
    /** Children to render in the card header. */
    children?: Snippet
    /** Children to render in the card footer. */
    footer?: Snippet
  }

  let {
    value = $bindable(),
    close,
    dir = false,
    file = false,
    height = '100%',
    children,
    footer,
  }: Props = $props()

  let filter = $state('')
  const fb = new FileBrowser(value, v => ((value = v), close()))
  const filt = $derived(filter.toLowerCase())
  const dirs = $derived(fb.wd.dirs?.filter(d => d.toLowerCase().includes(filt)) || [])
  const files = $derived(fb.wd.files?.filter(f => f.toLowerCase().includes(filt)) || [])
  const dirsCount = $derived(fb.wd.dirs?.length ?? 0)
  const fileCount = $derived(fb.wd.files?.length ?? 0)
</script>

<Card style="height: {height};min-height: 400px;">
  <CardHeader>
    <!-- Path title (input group). -->
    <form onsubmit={e => fb.cd(e, fb.input, true)}>
      <InputGroup>
        <Button
          outline
          onclick={e => fb.cd(e, fb.wd.mom || fb.wd.sep, true)}
          disabled={fb.loading}
          type="button">
          {#if fb.loading}
            <!-- Loading spinner. -->
            <Fa i={faSpinner} c1="steelblue" d2="firebrick" scale={1.5} spin />
          {:else}
            <!-- Up button. -->
            <Fa i={faArrowUpToArc} c1="steelblue" d2="firebrick" scale={1.5} />
          {/if}
        </Button>
        <InputGroupText><T id="LogFiles.titles.Path" /></InputGroupText>
        <Input bind:value={fb.input} />
        <!-- Go button. -->
        {#if fb.input !== fb.wd.path}
          <Button type="submit" color="primary" outline><T id="FileBrowser.Go" /></Button>
        {/if}
        <!-- Select path button. -->
        {#if !file}
          <Button
            id="fBut"
            color="success"
            outline
            type="button"
            onclick={e => fb.select(e, fb.input, true)}>
            <Fa i={faCheck} c1="limegreen" d1="green" scale={1.5} />
          </Button>
          <Tooltip target="fBut">
            <T id="FileBrowser.SelectPath" path={fb.wd.path} /></Tooltip>
        {/if}
      </InputGroup>
    </form>

    <!-- Children (a description of the file browser). -->
    {@render children?.()}
  </CardHeader>

  <CardBody class="overflow-auto h-100 p-0">
    <!-- Error message. -->
    {#if fb.respErr}
      <Card outline color="danger" class="m-2 text-center" body>{fb.respErr}</Card>
    {/if}
    <FileList {fb} {dir} {dirs} {files} />
  </CardBody>

  <CardFooter>
    <ActionBar bind:filter {fb} />
    <ul class="d-inline-block mb-0 ps-2">
      {#if !filt}
        <li><T id="FileBrowser.Folders" count={dirsCount} /></li>
        <li><T id="FileBrowser.Files" count={fileCount} /></li>
      {:else}
        <li>
          <T id="FileBrowser.FoldersFiltered" count={dirsCount} filtered={dirs.length} />
        </li>
        <li>
          <T id="FileBrowser.FilesFiltered" count={fileCount} filtered={files.length} />
        </li>
      {/if}
      {#if value}<li><T id="FileBrowser.Selected" path={value} /></li>{/if}
    </ul>
    {@render footer?.()}
  </CardFooter>
</Card>
