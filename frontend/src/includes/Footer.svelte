<!-- Footer for any page that saves configuration.
  - Shows a save button.
  - Shows a status message.
  - Shows a success message.
  - Shows a form error message.
  - Shows a form error message.
  - TIED to profile.svelte.ts.
-->
<script lang="ts">
  import { Alert, Button, CardFooter, Col, Row, Spinner } from '@sveltestrap/sveltestrap'
  import { profile } from '../api/profile.svelte'
  import T, { _, time } from './Translate.svelte'
  import { age } from './util'

  type Props = {
    /** The save button runs this function when clicked. */
    submit: (e: Event) => void
    /** The text to display when the save button is clicked. Must be translation key. */
    successText?: string
    /** The text to display on the save button. Must be translation key. */
    saveButtonText?: string
    /** The optional description of the save button. Must be translation key. */
    saveButtonDescription?: string
    /** Whether the save button is disabled. */
    saveDisabled?: boolean
    /** The children to render. */
    children?: () => any
  }

  let {
    submit,
    successText = 'phrases.ConfigurationSaved',
    saveButtonText = 'buttons.SaveConfiguration',
    saveButtonDescription = '',
    saveDisabled = false,
    children = undefined,
  }: Props = $props()

  let submitting = $state(false)
  let successTime = $state(new Date())
  let submitted = $state(false)

  async function onclick(e: Event) {
    e.preventDefault()
    submitting = true
    submitted = false
    await submit(e)
    submitting = false
    submitted = true
    successTime = new Date()
    profile.now = successTime.getTime() // speed up the timer display
  }

  function toggle() {
    submitted = false
    if (profile.status) profile.clearStatus()
  }

  // These are derived values that are used to display the status messages.
  let color = $derived(
    profile.formError ? 'danger' : profile.status ? 'warning' : 'success',
  )
  let isOpen = $derived(!!(profile.formError || profile.status || submitted))
  let disabled = $derived(submitting || saveDisabled)
  const values = $derived({
    timeDuration: age(profile.now - new Date(successTime ?? new Date()).getTime()),
  })
  let msg = $derived(
    profile.formError || profile.status || (submitted && $_(successText, { values })),
  )
</script>

<CardFooter>
  <div class="footer pb-2">
    <Row>
      <Col style="max-width: fit-content;" class="mt-1">
        <!-- Save Button -->
        <Button size="lg" color="notifiarr" type="submit" {disabled} {onclick}>
          {#if submitting}<Spinner size="sm" />{/if}
          {submitting ? $_('phrases.SavingConfiguration') : $_(saveButtonText)}
        </Button>
        <!-- Save Button Description -->
        {#if saveButtonDescription}
          <br /><small class="ms-2 text-muted">{$_(saveButtonDescription)}</small>
        {/if}
      </Col>

      <!-- Status Message, goes beside button -->
      <Col>
        <Alert {color} {toggle} {isOpen} closeClassName="close" class="submit-alert">
          {msg}
        </Alert>
      </Col>
    </Row>
  </div>
  {@render children?.()}
</CardFooter>

<style>
  /* The next two are used by the alerts that pop up when you click a save button.
   * This makes the alert match the height of a lg button.
   */
  .footer :global(.close) {
    position: absolute !important;
    right: 0;
    top: -5px !important;
  }

  .footer :global(.submit-alert) {
    margin: 2px 0px 0px 0;
    padding: 10px 33px 10px 10px !important;
    min-height: 50px;
    max-height: fit-content;
    font-size: 18px;
  }
</style>
