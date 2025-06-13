<script lang="ts">
  import Input from '../../includes/Input.svelte'
  import { Col, Row } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import { type ServiceConfig } from '../../api/notifiarrConfig'
  import type { ChildProps } from '../../includes/Instances.svelte'
  import { profile } from '../../api/profile.svelte'
  import { slide } from 'svelte/transition'
  import type { SvelteComponent } from 'svelte'
  import Http from './HTTP.svelte'
  import Proc from './Process.svelte'
  import Ping from './Ping.svelte'
  import Tcp from './TCP.svelte'

  let {
    form = $bindable(),
    original,
    app,
    index,
    validate,
  }: ChildProps<ServiceConfig> = $props()

  // When the check type changes, reset the form.
  const onchange = () => {
    if (form.type === original.type) form = { ...original }
    else form.expect = form.value = ''
    reset()
  }

  let pages = $state<Record<string, SvelteComponent<ChildProps<ServiceConfig>> | null>>({
    http: null,
    process: null,
    ping: null,
    icmp: null,
    tcp: null,
  })

  // This is called by Instances.svelte when the reset button is clicked.
  // Calls the exported reset method of the current page component.
  export const reset = () => pages[form.type]?.reset?.()
</script>

<div class="serviceCheck">
  <Row>
    <Col md={6}>
      <Input
        id={app.id + '.name'}
        bind:value={form.name}
        original={original.name}
        {validate} />
    </Col>
    <Col md={6}>
      <Input
        type="select"
        id={app.id + '.type'}
        bind:value={form.type}
        original={original?.type}
        {onchange}
        {validate}
        options={['process', 'http', 'tcp', 'ping', 'icmp'].map(type => ({
          name: $_(`ServiceChecks.type.options.${type}`),
          value: type,
          disabled: type === 'ping' && $profile.isWindows,
        }))} />
    </Col>
  </Row>

  {#if form.type === 'http'}
    <div class="row" transition:slide>
      <Http {form} {original} {app} {index} {validate} bind:this={pages.http} />
    </div>
  {:else if form.type === 'process'}
    <div class="row" transition:slide>
      <Proc {form} {original} {app} {index} {validate} bind:this={pages.process} />
    </div>
  {:else if form.type === 'icmp'}
    <div class="row" transition:slide>
      <Ping {form} {original} {app} {index} {validate} bind:this={pages.icmp} />
    </div>
  {:else if form.type === 'ping'}
    <div class="row" transition:slide>
      <Ping {form} {original} {app} {index} {validate} bind:this={pages.ping} />
    </div>
  {:else if form.type === 'tcp'}
    <div class="row" transition:slide>
      <Tcp {form} {original} {app} {index} {validate} bind:this={pages.tcp} />
    </div>
  {/if}

  <Row>
    <Col md={6}>
      <Input
        id={app.id + '.timeout'}
        type="timeout"
        bind:value={form.timeout}
        original={original?.timeout}
        {validate} />
    </Col>
    <Col md={6}>
      <Input
        id={app.id + '.interval'}
        type="interval"
        bind:value={form.interval}
        original={original?.interval}
        {validate} />
    </Col>
  </Row>
</div>

<style>
  .serviceCheck :global(.changed) {
    background-color: rgba(205, 92, 92, 0.322) !important;
  }
</style>
