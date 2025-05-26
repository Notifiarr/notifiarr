<script lang="ts" module>
  import type { App } from './formsTracker.svelte'
  /** These are the props that are passed to the child component. */
  export type ChildProps<T> = {
    form: T
    original: T
    app: App
    validate?: (id: string, value: any) => string
    index?: number
    reset?: () => void
  }
</script>

<script lang="ts">
  import {
    Accordion,
    AccordionHeader,
    Badge,
    Button,
    Card,
  } from '@sveltestrap/sveltestrap'
  import T, { _ } from './Translate.svelte'
  import { slide } from 'svelte/transition'
  import { deepEqual } from './util'
  import InstanceHeader from './InstanceHeader.svelte'
  import type { FormListTracker } from './formsTracker.svelte'
  import type { Component, Snippet } from 'svelte'

  let {
    flt,
    Child,
    headerActive = $bindable(),
    headerCollapsed = $bindable(),
    deleteButton = 'phrases.DeleteInstance',
  }: {
    flt: FormListTracker
    Child: Component<ChildProps<any>>
    headerActive: Snippet<[number]>
    headerCollapsed?: Snippet<[number]>
    deleteButton?: string
  } = $props()
</script>

<InstanceHeader {flt} />

{#if flt.instances.length > 0}
  <div class="instances" transition:slide>
    <Accordion class="mb-2">
      {#each flt.instances as instance, index}
        {@const changed = !deepEqual(instance, flt.original[index] ?? flt.app.empty)}
        <div class="accordion-item">
          <AccordionHeader
            onclick={() => (flt.active = index)}
            class={flt.active !== index ? 'collapsed d-block' : ''}>
            <h5 class="mb-0">
              {@render headerActive(index)}
              {#if !flt.isValid(index)}
                <Badge color="danger"><T id="phrases.Invalid" /></Badge>
              {:else if index + 1 > flt.original.length}
                <Badge color="info"><T id="phrases.New" /></Badge>
              {:else if changed}
                <Badge color="warning"><T id="phrases.Changed" /></Badge>
              {/if}
            </h5>
            {#if flt.active !== index}
              <div style="overflow: clip;">
                <span class="text-muted fs-6 mt-0 header-collapsed">
                  {@render headerCollapsed?.(index)}
                </span>
              </div>
            {/if}
          </AccordionHeader>

          {#key flt.active}
            <Card
              class="accordion-collapse {flt.active === index ? 'd-block' : 'd-none'}">
              <div class="accordion-body" transition:slide={{ duration: 350, axis: 'y' }}>
                <Child
                  bind:form={flt.instances[index]!}
                  original={flt.original[index] ?? flt.app.empty}
                  app={flt.app}
                  validate={(id, value) => flt.validate(id, value, index)}
                  {index} />
                <Button
                  color="danger"
                  class="float-end"
                  outline
                  onclick={async () => await flt.delInstance(index)}>
                  {$_(deleteButton)}
                </Button>
                <Button
                  color="primary"
                  class="float-end me-2"
                  outline
                  disabled={!changed}
                  onclick={() => flt.resetForm(index)}>
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

<Button color="success" outline onclick={flt.addInstance}>
  {$_(flt.app.id + '.addInstance')}
</Button>

<style>
  .instances :global(.accordion .badge) {
    position: absolute;
    right: 60px;
    border-radius: 12px;
    text-transform: uppercase;
  }

  /* This allows the sub header to span the entire length of the accordion. */
  .header-collapsed {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: initial;
    word-break: break-all;
    display: -webkit-box;
    line-clamp: 1;
    -webkit-line-clamp: 1;
    -webkit-box-orient: vertical;
  }
</style>
