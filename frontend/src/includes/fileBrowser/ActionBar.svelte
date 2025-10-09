<script lang="ts">
  import {
    InputGroup,
    InputGroupText,
    Input,
    Dropdown,
    DropdownToggle,
    DropdownMenu,
    DropdownItem,
    Modal,
    ModalBody,
    Button,
    ModalHeader,
  } from '@sveltestrap/sveltestrap'
  import Fa from '../Fa.svelte'
  import { faArrowUpFromBracket } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import T from '../Translate.svelte'
  import type { FileBrowser } from './browser.svelte'
  import { theme } from '../theme.svelte'
  import { faSpinner } from '@fortawesome/sharp-duotone-regular-svg-icons'

  type Props = { filter: string; fb: FileBrowser }
  let { filter = $bindable(), fb }: Props = $props()
  let newPath = $state('')
  let newType = $state('')
  let isOpen = $state(false)

  const cancel = () => ((newPath = ''), (isOpen = false))
  const open = (type: string) => ((newType = type), (isOpen = true))
  const create = async (e: Event, path: string, dir: boolean) => {
    e.preventDefault()
    await fb.create(path, dir)
    isOpen = false
    newPath = newType = ''
  }
</script>

<InputGroup class="mb-2">
  <InputGroupText><T id="Actions.titles.Filter" /></InputGroupText>
  <Input bind:value={filter} />
  {#if fb.wd.path}
    <Dropdown group dropup>
      <DropdownToggle outline class="rounded-0 rounded-end" disabled={!fb.wd.path}>
        <Fa i={faArrowUpFromBracket} c1="gray" d1="gainsboro" c2="orange" scale={1.5} />
      </DropdownToggle>
      <DropdownMenu>
        <DropdownItem onclick={() => open('CreateFolder')}>
          <T id="FileBrowser.CreateFolder" /></DropdownItem>
        <DropdownItem onclick={() => open('CreateFile')}>
          <T id="FileBrowser.CreateFile" /></DropdownItem>
      </DropdownMenu>
    </Dropdown>
  {/if}
</InputGroup>

<!-- New folder / file modal. Path input. -->
<Modal bind:isOpen theme={$theme} centered contentClassName="border-warning-subtle">
  <ModalHeader toggle={cancel}><T id="FileBrowser.{newType}" /></ModalHeader>
  <ModalBody>
    <T id="FileBrowser.{newType}In" path={fb.wd.path} />
    <form onsubmit={e => create(e, newPath, newType == 'CreateFolder')}>
      <InputGroup>
        <Input bind:value={newPath} />
        <Button color="success" outline type="submit" disabled={!newPath || fb.loading}>
          {#if fb.loading}
            <Fa i={faSpinner} c1="gray" d1="gainsboro" c2="orange" scale={1.5} spin />
            <T id="FileBrowser.Creating" />
          {:else}
            <T id="buttons.Create" />
          {/if}
        </Button>
      </InputGroup>
    </form>
  </ModalBody>
</Modal>
