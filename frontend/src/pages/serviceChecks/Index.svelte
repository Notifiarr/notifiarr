<script lang="ts" module>
  import { empty, merge, page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import { CardBody, Col, Row } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import { FormListTracker } from '../../includes/formsTracker.svelte'
  import { profile } from '../../api/profile.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import type { ServiceConfig } from '../../api/notifiarrConfig'
  import { get } from 'svelte/store'
  import type { App } from '../../includes/formsTracker.svelte'
  import { type IconSource } from '../../includes/Fa.svelte'
  import Pulse from 'phosphor-svelte/lib/Pulse'
  import Globe from 'phosphor-svelte/lib/Globe'
  import Cpu from 'phosphor-svelte/lib/Cpu'
  import PingPong from 'phosphor-svelte/lib/PingPong'
  import PlugsConnected from 'phosphor-svelte/lib/PlugsConnected'
  import Instances from '../../includes/Instances.svelte'
  import Check from './Check.svelte'
  import { validator as httpValidator } from './HTTP.svelte'
  import { validator as processValidator } from './Process.svelte'
  import { validator as pingValidator } from './Ping.svelte'
  import { validator as tcpValidator } from './TCP.svelte'
  import Fa from '../../includes/Fa.svelte'
  import { deepEqual } from '../../includes/util'
  import Input from '../../includes/Input.svelte'

  // Local state that syncs with profile store.
  let config = $state($profile.config)

  const submit = async () => {
    await profile.writeConfig({ ...config, service: flt.instances })
    if (profile.error) return
    config = $profile.config
    flt.resetAll() // clears the delete counters.
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
      return val ? found : get(_)('phrases.NameMustNotBeEmpty')
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
    envPrefix: 'SERVICE',
    logo: Pulse,
    iconProps: { c1: 'coral', c2: 'lightcoral' },
    disabled: [],
    hidden: [],
    empty,
    merge,
    validator,
  }

  let flt = $derived(new FormListTracker($profile.config.service ?? [], app))

  $effect(() => {
    nav.formChanged = !deepEqual($profile.config, config) || flt.formChanged
  })

  // Shown next to the check name in each accordion header.
  const icons: Record<string, IconSource> = {
    http: Globe,
    process: Cpu,
    ping: PingPong,
    icmp: PingPong,
    tcp: PlugsConnected,
  }
</script>

<Header {page}>
  {#snippet description()}
    <Row>
      <Col xxl={9} xl={8} md={7} sm={6} xs={12}>
        {@html $_('navigation.pageDescription.' + page.id)}
      </Col>
      <Col xxl={3} xl={4} md={5} sm={6} xs={12}>
        <div class="d-block d-sm-none"><hr /></div>
        <Input
          class="mb-0"
          id="config.services.disabled"
          envVar="SERVICES_DISABLED"
          type="select"
          bind:value={config.services!.disabled}
          original={$profile.config.services?.disabled} />
      </Col>
    </Row>
  {/snippet}
</Header>

<CardBody>
  <!-- Services Section -->
  <Instances {flt} Child={Check} deleteButton={app.id + '.DeleteCheck'}>
    {#snippet headerActive(index)}
      <Fa
        flip={flt.original?.[index]?.type === 'icmp'}
        i={icons[flt.original?.[index]?.type]}
        c1="#0E6655"
        c2="#0B5345"
        d1="#9FE2BF"
        d2="#40E0D0"
        logo>
        {index + 1}. {flt.original?.[index]?.name}
      </Fa>
    {/snippet}
    {#snippet headerCollapsed(index)}
      {$_('ServiceChecks.type.options.' + flt.original?.[index]?.type)}:
      {flt.original?.[index]?.value.split('|')[0]}
    {/snippet}
  </Instances>
</CardBody>

<Footer {submit} saveDisabled={!nav.formChanged || flt.invalid} />
