<script lang="ts" module>
  import { empty, merge, page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import { CardBody } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import { profile } from '../../api/profile.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import type { ServiceConfig } from '../../api/notifiarrConfig'
  import { get } from 'svelte/store'
  import type { App } from '../../includes/formsTracker.svelte'
  import { faStaffSnake } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Instances from '../../includes/Instances.svelte'
  import Check from './Check.svelte'
  import { validator as httpValidator } from './HTTP.svelte'
  import { validator as processValidator } from './Process.svelte'
  import { validator as pingValidator } from './Ping.svelte'
  import { validator as tcpValidator } from './TCP.svelte'

  const submit = async () => {
    await profile.writeConfig({ ...$profile.config, service: flt.instances })
    if (!profile.error) flt.resetAll() // clears the delete counters.
  }

  const validator = (id: string, val: any, idx: number, c: ServiceConfig[]): string => {
    id = id.split('.').pop() ?? id

    if (id == 'name') {
      let found = ''
      c?.forEach((m, i) => {
        if (i !== idx && m?.name === val) {
          found = get(_)('phrases.NameInUseByInstance', { values: { number: i + 1 } })
          return
        }
      })
      if (found) return found
      return val ? '' : get(_)('phrases.NameMustNotBeEmpty')
    } else if (c?.[idx]?.type === 'http') {
      return httpValidator(id, val)
    } else if (c?.[idx]?.type === 'process') {
      return processValidator(id, val)
    } else if (['ping', 'icmp'].includes(c?.[idx]?.type)) {
      return pingValidator(id, val)
    } else if (c?.[idx]?.type === 'tcp') {
      return tcpValidator(id, val)
    } else {
      return ''
    }
  }

  const app: App<ServiceConfig> = {
    name: 'Checks',
    id: 'ServiceChecks',
    logo: faStaffSnake,
    iconProps: { c1: 'coral', c2: 'lightcoral' },
    disabled: [],
    hidden: [],
    empty,
    merge,
    validator,
  }

  let flt = $derived(new FormListTracker($profile.config.service ?? [], app))

  $effect(() => {
    nav.formChanged = flt.formChanged
  })
</script>

<Header {page} />

<CardBody>
  <Instances {flt} Child={Check} deleteButton={app.id + '.DeleteCheck'}>
    {#snippet headerActive(index)}
      {index + 1}. {flt.original?.[index]?.name}
    {/snippet}
    {#snippet headerCollapsed(index)}
      {flt.original?.[index]?.type}: {flt.original?.[index]?.value.split('|')[0]}
    {/snippet}
  </Instances>
</CardBody>

<Footer {submit} saveDisabled={!flt.formChanged || flt.invalid} />
