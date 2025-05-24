<script lang="ts">
  import {
    Accordion,
    AccordionHeader,
    Badge,
    Button,
    Card,
  } from '@sveltestrap/sveltestrap'
  import Instance from './Instance.svelte'
  import T, { _ } from './Translate.svelte'
  import { slide } from 'svelte/transition'
  import { deepEqual } from './util'
  import InstanceHeader from './InstanceHeader.svelte'
  import type { InstanceFormValidator } from './instanceFormValidator.svelte'

  let { iv }: { iv: InstanceFormValidator } = $props()
</script>

<InstanceHeader {iv} />

{#if iv.instances.length > 0}
  <div class="instances" transition:slide>
    <Accordion class="mb-2">
      {#each iv.instances as instance, index}
        {@const changed = !deepEqual(instance, iv.original[index] ?? iv.app.empty)}
        <div class="accordion-item">
          <AccordionHeader
            onclick={() => (iv.active = index)}
            class={iv.active !== index ? 'collapsed d-block' : ''}>
            <h5 class="mb-0">
              {index + 1}. {iv.original[index]?.name}
              {#if !iv.isValid(index)}
                <Badge color="danger"><T id="phrases.Invalid" /></Badge>
              {:else if index + 1 > iv.original.length}
                <Badge color="info"><T id="phrases.New" /></Badge>
              {:else if changed}
                <Badge color="warning"><T id="phrases.Changed" /></Badge>
              {/if}
            </h5>
            {#if iv.active !== index}
              <span class="text-muted fs-6 mt-0">
                {iv.original[index]?.url}{iv.original[index]?.host}
              </span>
            {/if}
          </AccordionHeader>

          {#key iv.active}
            <Card class="accordion-collapse {iv.active === index ? 'd-block' : 'd-none'}">
              <div class="accordion-body" transition:slide={{ duration: 350, axis: 'y' }}>
                <Instance
                  bind:form={iv.instances[index]!}
                  original={iv.original[index] ?? iv.app.empty}
                  app={iv.app}
                  validate={(id, value) => iv.validate(id, value, index)}
                  {index} />
                <Button
                  color="danger"
                  class="float-end"
                  outline
                  onclick={async () => await iv.delInstance(index)}>
                  {$_('phrases.DeleteInstance')}
                </Button>
                <Button
                  color="primary"
                  class="float-end me-2"
                  outline
                  disabled={!changed}
                  onclick={() => iv.resetForm(index)}>
                  {$_('buttons.ResetForm')}
                </Button>
                <div style="clear: both;"></div>
              </div>
            </Card>
          {/key}
        </div>
      {/each}
    </Accordion>
  </div>
{/if}

<Button color="success" outline onclick={iv.addInstance}>
  {$_(iv.app.id + '.addInstance')}
</Button>

<style>
  .instances :global(.accordion .badge) {
    position: absolute;
    right: 60px;
    border-radius: 12px;
    text-transform: uppercase;
  }
</style>
