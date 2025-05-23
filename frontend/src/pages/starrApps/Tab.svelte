<script lang="ts">
  import { Starr } from './starr.svelte'
  import type { App, Form } from '../../includes/Instance.svelte'
  import Fa from '../../includes/Fa.svelte'
  import { faExclamationTriangle } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { TabPane } from '@sveltestrap/sveltestrap'
  import Instances from '../../includes/Instances.svelte'

  type Props = {
    app: App
    equal: boolean
    original: (Form | null)[]
    valid: boolean
    validate: (id: string, index: number, value: any, reset?: boolean) => string
    form: Instances | undefined
    instances: (Form | null)[]
    removed: {
      Sonarr: number[]
      Radarr: number[]
      Readarr: number[]
      Lidarr: number[]
      Prowlarr: number[]
    }
    tab: string
  }

  let {
    app,
    equal,
    original,
    valid,
    validate,
    form = $bindable(),
    instances = $bindable(),
    removed = $bindable(),
    tab = $bindable(),
  }: Props = $props()
</script>

<TabPane
  tabId={app.name.toLowerCase()}
  active={tab === app.name.toLowerCase()}
  onclick={() => (tab = app.name.toLowerCase())}>
  <div slot="tab">
    <span class={!valid ? 'text-danger' : ''}>
      {Starr.title[app.name as keyof typeof Starr.title]}
    </span>
    {#if !valid}
      <Fa i={faExclamationTriangle} c1="red" c2="red" />
    {:else if !equal || removed[app.name as keyof typeof removed].length > 0}
      <Fa i={faExclamationTriangle} c1="orange" c2="orange" />
    {/if}
  </div>

  <Instances
    {validate}
    remove={index => removed[app.name as keyof typeof removed].push(index)}
    bind:instances
    bind:this={form}
    {original}
    {app} />
</TabPane>
