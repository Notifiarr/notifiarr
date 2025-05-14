<script lang="ts">
  import { Alert, Button, CardFooter, Col, Row } from '@sveltestrap/sveltestrap'
  import { profile } from '../api/profile.svelte'
  import T, { _ } from '../lib/Translate.svelte'
  import { age } from '../lib/util'

  export let submit: (e: Event) => void
  export let successText: string = 'phrases.ConfigurationSaved'
  export let saveButtonText: string = 'buttons.SaveConfiguration'
  export let saveButtonDescription: string = ''
  export let saveDisabled: boolean = false
</script>

<CardFooter>
  <div class="footer">
    <Row>
      <Col style="max-width: fit-content;">
        <!-- Save Button -->
        <Button
          size="lg"
          color="primary"
          type="submit"
          class="mt-1"
          disabled={profile.status !== '' || saveDisabled}
          onclick={(e: Event) => (e.preventDefault(), submit(e))}>
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
  <slot />
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
