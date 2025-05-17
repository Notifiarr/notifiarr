<script lang="ts">
  import { Alert, Button, CardFooter, Col, Row } from '@sveltestrap/sveltestrap'
  import { profile } from '../api/profile.svelte'
  import T, { _ } from './Translate.svelte'
  import { age } from './util'

  type Props = {
    /** The save button runs this function when clicked. */
    submit: (e: Event) => void
    /** The text to display when the save button is clicked. Must be translation key. */
    successText?: string
    /** The text to display on the save button. Must be translation key. */
    saveButtonText?: string
    /** The optionaldescription of the save button. Must be translation key. */
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

  async function onclick(e: Event) {
    e.preventDefault()
    await submit(e)
  }
</script>

<CardFooter>
  <div class="footer pb-2">
    <Row>
      <Col style="max-width: fit-content;">
        <!-- Save Button -->
        <Button
          size="lg"
          color="primary"
          type="submit"
          class="mt-1"
          disabled={profile.status !== '' || saveDisabled}
          {onclick}>
          {profile.status ? $_('phrases.SavingConfiguration') : $_(saveButtonText)}
        </Button>
        <!-- Save Button Description -->
        {#if saveButtonDescription}
          <br /><small class="ms-2 text-muted">{$_(saveButtonDescription)}</small>
        {/if}
      </Col>
      <!-- Status Message, goes beside button -->
      <Col>
        <Alert
          color={profile.error ? 'danger' : profile.success ? 'success' : 'warning'}
          toggle={profile.status ? undefined : () => profile.clearStatus()}
          isOpen={!!(profile.error || profile.status || profile.success)}
          closeClassName="submit-alert-close"
          class="submit-alert">
          <!-- These happen in order, and only one will display something at a time.-->
          {profile.status}
          {profile.error}
          {#if profile.success}
            <T id={successText} age={age(profile.successAge, true)} />
          {/if}
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
  .footer :global(.submit-alert-close) {
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
