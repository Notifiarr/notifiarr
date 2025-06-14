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
  import { faMicrochip, faStaffSnake } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import {
    faGlobePointer,
    faHexagonNodesBolt,
    faPingPongPaddleBall,
  } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { type IconDefinition } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import Instances from '../../includes/Instances.svelte'
  import Check from './Check.svelte'
  import { validator as httpValidator } from './HTTP.svelte'
  import { validator as processValidator } from './Process.svelte'
  import { validator as pingValidator } from './Ping.svelte'
  import { validator as tcpValidator } from './TCP.svelte'
  import Fa from '../../includes/Fa.svelte'

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

  // Shown next to the check name in each accordion header.
  const icons: Record<string, IconDefinition> = {
    http: faGlobePointer,
    process: faMicrochip,
    ping: faPingPongPaddleBall,
    icmp: faPingPongPaddleBall,
    tcp: faHexagonNodesBolt,
  }
</script>

<Header {page} />

<CardBody>
  <Instances {flt} Child={Check} deleteButton={app.id + '.DeleteCheck'}>
    {#snippet headerActive(index)}
      <Fa
        flip={flt.original?.[index]?.type === 'icmp' ? 'horizontal' : undefined}
        i={icons[flt.original?.[index]?.type]}
        c1="#0E6655"
        c2="#0B5345"
        d1="#9FE2BF"
        d2="#40E0D0"
        scale="1.8"
        class="header-icon" />
      {index + 1}. {flt.original?.[index]?.name}
    {/snippet}
    {#snippet headerCollapsed(index)}
      {$_('ServiceChecks.type.options.' + flt.original?.[index]?.type)}:
      {flt.original?.[index]?.value.split('|')[0]}
    {/snippet}
  </Instances>
</CardBody>

<Footer {submit} saveDisabled={!flt.formChanged || flt.invalid} />

<style>
  :global(.header-icon) {
    margin-right: 8px;
    margin-bottom: 4px;
  }
</style>
