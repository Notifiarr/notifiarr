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
    ModalFooter,
  } from '@sveltestrap/sveltestrap'
  import Fa from '../Fa.svelte'
  import { faArrowUpFromBracket } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import T from '../Translate.svelte'
  import type { FileBrowser } from './browser.svelte'
  import { theme } from '../theme.svelte'

  type Props = { filter: string; fb: FileBrowser }
  let { filter = $bindable(), fb }: Props = $props()
  let newPath = $state('')
  let newType = $state('')
  let isOpen = $state(false)

  function create() {
    isOpen = false
    newPath = ''
    newType = 'CreateFolder'
  }
  const cancel = () => ((newPath = ''), (isOpen = false))
  const open = (type: string) => ((newType = type), (isOpen = true))
</script>

<InputGroup class="mb-2">
  <InputGroupText><T id="Actions.titles.Filter" /></InputGroupText>
  <Input bind:value={filter} />
  <Dropdown group dropup>
    <DropdownToggle outline class="rounded-0 rounded-end">
      <Fa i={faArrowUpFromBracket} c1="gray" d1="gainsboro" c2="orange" scale={1.5} />
    </DropdownToggle>
    <DropdownMenu>
      <DropdownItem onclick={() => open('CreateFolder')}>
        <T id="FileBrowser.CreateFolder" /></DropdownItem>
      <DropdownItem onclick={() => open('CreateFile')}>
        <T id="FileBrowser.CreateFile" /></DropdownItem>
    </DropdownMenu>
  </Dropdown>
</InputGroup>

<!-- New folder / file modal.-->
<Modal bind:isOpen theme={$theme} centered contentClassName="border-warning-subtle">
  <ModalHeader toggle={cancel}><T id="FileBrowser.{newType}" /></ModalHeader>
  <ModalBody>
    <T id="FileBrowser.{newType}In" path={fb.wd.path} />
    <InputGroup>
      <Input bind:value={newPath} />
      <Button color="info" outline onclick={create}><T id="buttons.Create" /></Button>
    </InputGroup>
  </ModalBody>
  <ModalFooter>This does not work yet. it will next</ModalFooter>
</Modal>
