<script lang="ts">
  import {
    Accordion,
    AccordionHeader,
    Badge,
    Button,
    Card,
  } from '@sveltestrap/sveltestrap'
  import Instance, { type Form, type App } from './Instance.svelte'
  import T, { _ } from './Translate.svelte'
  import { slide } from 'svelte/transition'
  import { deepEqual, delay, success } from './util'
  import InstanceHeader from './InstanceHeader.svelte'

  type Props = {
    instances: (Form | null)[]
    original: (Form | null)[]
    app: App
    remove: (index: number) => void
    validate: (id: string, index: number, value: any, reset?: boolean) => string
  }
  let { instances = $bindable(), original, app, remove, validate }: Props = $props()

  let active: number | undefined = $state(0)
  let accordions = $state<Instance[]>([])
  let deleted = $state<(Form | null)[]>([])

  const addInstance = () => (instances.push(app.empty), (active = instances.length - 1))
  const delInstance = async (index: number) => {
    active = undefined
    await delay(400)
    instances.splice(index, 1)
    accordions[index].resetFeedback()
    remove(index)
    if (index < original.length) {
      deleted.push(...original.splice(index, 1))
      await delay(100)
      success($_('phrases.ItsGone'))
    }
    active = 0
  }

  export const clear = () => (deleted = [])
  const formChanged = $derived(deleted.length > 0 || !deepEqual(instances, original))
</script>

<InstanceHeader {app} deleted={deleted.length} changed={formChanged} />

{#if instances && instances.length > 0}
  <div class="instances" transition:slide>
    <Accordion class="mb-2">
      {#each instances as instance, index}
        {@const changed = !deepEqual(instance, original[index] ?? app.empty)}
        <div class="accordion-item">
          <AccordionHeader
            onclick={() => (active = index)}
            class={active !== index ? 'collapsed d-block' : ''}>
            <h5 class="mb-0">
              {index + 1}. {original[index]?.name}
              {#if !(accordions[index]?.valid() ?? true)}
                <Badge color="danger"><T id="phrases.Invalid" /></Badge>
              {:else if index + 1 > original.length}
                <Badge color="info"><T id="phrases.New" /></Badge>
              {:else if changed}
                <Badge color="warning"><T id="phrases.Changed" /></Badge>
              {/if}
            </h5>
            {#if active !== index}
              <span class="text-muted fs-6 mt-0">
                {original[index]?.url}{original[index]?.host}
              </span>
            {/if}
          </AccordionHeader>

          {#key active}
            <Card class="accordion-collapse {active === index ? 'd-block' : 'd-none'}">
              <div class="accordion-body" transition:slide={{ duration: 350, axis: 'y' }}>
                <Instance
                  bind:this={accordions[index]}
                  bind:form={instances[index]!}
                  original={original[index] ?? app.empty}
                  resetButton={false}
                  {app}
                  {validate}
                  {index} />
                <Button
                  color="danger"
                  class="float-end"
                  outline
                  onclick={async () => await delInstance(index)}>
                  {$_('phrases.DeleteInstance')}
                </Button>
                <Button
                  color="primary"
                  class="float-end me-2"
                  outline
                  disabled={!changed}
                  onclick={() => accordions[index].reset()}>
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

<Button color="success" outline onclick={addInstance}>
  {$_(app.id + '.addInstance')}
</Button>

<style>
  .instances :global(.accordion .badge) {
    position: absolute;
    right: 60px;
    border-radius: 12px;
    text-transform: uppercase;
  }
</style>
