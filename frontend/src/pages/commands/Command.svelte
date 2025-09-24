<script lang="ts">
  import Input from '../../includes/Input.svelte'
  import {
    Button,
    Col,
    Row,
    Input as Box,
    ModalHeader,
    ModalBody,
    ModalFooter,
    FormGroup,
    Badge,
  } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import { type Command } from '../../api/notifiarrConfig'
  import type { ChildProps } from '../../includes/Instances.svelte'
  import CheckedInput from '../../includes/CheckedInput.svelte'
  import { delay } from '../../includes/util'
  import Output from './Output.svelte'
  import MyModal from '../../includes/MyModal.svelte'

  let {
    form = $bindable(),
    original,
    app,
    index,
    validate,
    active = false,
  }: ChildProps<Command> = $props()

  /** Track our output component so we can call methods in it. */
  let output: Output | null = $state(null)
  /** If a command finishes quickly this updates its output after execution. */
  const refreshStats = async () => (await delay(1000), output?.getStats())
  // These are used to pass custom arguments into a command through a modal form.
  let runArgs = $state(form.argValues?.map(arg => '') ?? [])
  let formResolve: (() => void) | null = $state(null)
  let formReject: (() => void) | null = $state(null)
  /** Passed into Instances as a callback to get command arguments. */
  const getCommandArgs = async (): Promise<URLSearchParams> => {
    if (form.args === 0) {
      refreshStats() // do not await this.
      return new URLSearchParams()
    }
    await new Promise<void>((res, rej) => ((formResolve = res), (formReject = rej)))
    refreshStats() // do not await this.
    return new URLSearchParams(runArgs.map(arg => ['arg', arg]))
  }

  $effect(() => {
    // Reset the args when the form values change.
    if (active) runArgs = form.argValues?.map(a => '') ?? []
  })

  /** Checks the argValues (regex) against the given argument. */
  const checkRegex = (arg: string, i: number) => {
    let re = form.argValues?.[i] ?? ''
    if (!re.startsWith('^')) re = '^(' + re + ')$'
    return !!arg.match(new RegExp(re))
  }

  /** Closes the modal without running the command. */
  const cancel = (e?: Event) => {
    e?.preventDefault()
    formReject?.()
    formResolve = null
  }
</script>

<div class="command">
  <Row>
    <Col sm={7}>
      <Input
        id={app.id + '.name'}
        bind:value={form.name}
        original={original?.name}
        {validate} />
    </Col>
    <Col sm={5}>
      <Input
        type="select"
        id={app.id + '.notify'}
        bind:value={form.notify}
        original={original?.notify}
        {validate} />
    </Col>
    <Col md={12}>
      <CheckedInput
        id="command"
        params={getCommandArgs}
        {app}
        {index}
        bind:form
        bind:original
        {validate}
        disabled={form.command !== original?.command ||
          form.command == '' ||
          form.shell != original?.shell} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.log'}
        bind:value={form.log}
        original={original?.log}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="timeout"
        id={app.id + '.timeout'}
        bind:value={form.timeout}
        original={original?.timeout}
        noDisable
        {validate} />
    </Col>
  </Row>

  <!-- Show output component here. -->
  <Output {form} {index} {active} bind:this={output} />
</div>

<!-- Modal is used to run commands that have custom arguments (regexes). -->
<MyModal toggle={cancel} isOpen={formResolve != null} size="lg" centered>
  <ModalHeader><T id="Commands.enterCommandArguments" /></ModalHeader>
  <form onsubmit={e => e.preventDefault()}>
    <ModalBody>
      <p class="text-muted">
        <T id="Commands.commandRequiresArguments" count={form.args} />
      </p>
      <p><b class="text-primary font-monospace">{form.command}</b></p>
      {#each form.argValues ?? [] as arg, i}
        <FormGroup floating>
          <div slot="label">
            <Badge>{i + 1}</Badge> &nbsp; <b class="text-primary font-monospace">{arg}</b>
          </div>
          <Box
            tabindex={i + 1}
            bind:value={runArgs[i]}
            invalid={!checkRegex(runArgs[i], i)}
            feedback={$_('Commands.regexMismatch')} />
        </FormGroup>
      {/each}
    </ModalBody>
    <ModalFooter>
      <Button
        outline
        type="button"
        color="warning"
        onclick={cancel}
        tabindex={form.args + 3}>
        <T id="buttons.Cancel" /></Button>
      <Button
        type="submit"
        color="notifiarr"
        disabled={!runArgs.every(checkRegex)}
        tabindex={form.args + 2}
        onclick={e => {
          e.preventDefault()
          formResolve?.()
          formResolve = null
        }}><T id="buttons.Execute" /></Button>
    </ModalFooter>
  </form>
</MyModal>
