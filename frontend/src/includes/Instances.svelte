<script lang="ts">
  import { Accordion, AccordionItem } from '@sveltestrap/sveltestrap'
  import Instance, { type Form, type App } from './Instance.svelte'
  import { _ } from './Translate.svelte'

  type Props = { instances?: Form | Form[]; app: App }
  let { instances = $bindable(undefined), app }: Props = $props()
  let active = $state([] as Record<number, boolean>)
</script>

<h4>
  <img src={app.logo} alt="Logo" class="logo" />
  {$_(app.id + '.title')}
</h4>
<p>{@html $_(app.id + '.description')}</p>

{#if instances}
  {#if Array.isArray(instances)}
    <Accordion>
      {#each instances as instance, index}
        <AccordionItem active={active[index]} on:toggle={e => (active[index] = e.detail)}>
          <div slot="header">
            {#if active[index]}
              <!-- what you see in the header when it's picked. -->
              {index + 1}. {instance.name}
            {:else}
              <!-- what you see in the header when it's not picked (default view). -->
              {index + 1}. {instance.name} ({instance.url})
            {/if}
          </div>
          <Instance bind:form={instances[index]} {app} {index} />
        </AccordionItem>
      {/each}
    </Accordion>
  {:else}
    <Instance bind:form={instances} {app} index={0} />
  {/if}
{/if}

<style>
  .logo {
    height: 36px;
    margin-right: 6px;
    margin-left: -5px;
    padding-left: 0px;
    vertical-align: bottom;
    display: inline-block;
  }
</style>
