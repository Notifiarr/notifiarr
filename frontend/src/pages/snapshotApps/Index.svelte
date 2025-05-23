<script lang="ts" module>
  import { faCameraRetro } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'SnapshotApps',
    i: faCameraRetro,
    c1: 'burlywood',
    c2: 'darkgray',
    d1: 'burlywood',
    d2: 'silver',
  }
</script>

<script lang="ts">
  import { profile } from '../../api/profile.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import Instances from '../../includes/Instances.svelte'
  import mysqlLogo from '../../assets/logos/mysql.png'
  import nvidiaLogo from '../../assets/logos/nvidia.png'
  import type { MySQLConfig, NvidiaConfig } from '../../api/notifiarrConfig'
  import { deepCopy, deepEqual } from '../../includes/util'
  import Instance, { type Form } from '../../includes/Instance.svelte'
  import InstanceHeader from '../../includes/InstanceHeader.svelte'

  const validate = (id: string, index: number, value: any, reset?: boolean) => {
    if (reset) {
      invalid[index] = {}
      return ''
    }

    if (!invalid[index]) invalid[index] = {}
    invalid[index][id] = ''

    if (id.endsWith('.name')) {
      mysql.forEach((m, i) => {
        if (i !== index && m?.name === value) {
          invalid[index][id] = $_('phrases.NameInUseByInstance', {
            values: { number: i + 1 },
          })
          return
        }
      })
      if (value === '') invalid[index][id] = $_('phrases.NameMustNotBeEmpty')
    }

    if (id.endsWith('.host') && value === '') {
      invalid[index][id] = $_('phrases.HostMustNotBeEmpty')
    }

    if (id.endsWith('.username') && value === '') {
      invalid[index][id] = $_('phrases.UsernameMustNotBeEmpty')
    }

    //console.log('debug', id, value, invalid[index][id])
    return invalid[index][id]
  }

  let mysqlConfig: MySQLConfig = {
    name: '',
    host: '',
    username: '',
    password: '',
    timeout: '10s',
    interval: '5m0s',
  }

  let mysql = $state(deepCopy($profile.config.snapshot?.mysql ?? []))
  let nvidia = $state(deepCopy($profile.config.snapshot?.nvidia))
  let originalMysql = $derived(deepCopy($profile.config.snapshot?.mysql ?? []))
  let originalNvidia = $derived(deepCopy($profile.config.snapshot?.nvidia))
  let m = $state<Instances>()
  let invalid = $state<Record<number, Record<string, string>>>({})

  let formChanged = $derived(
    !deepEqual(mysql, originalMysql) || !deepEqual(nvidia, originalNvidia),
  )

  const mysqlApp = {
    name: 'MySQL',
    id: page.id + '.MySQL',
    logo: mysqlLogo,
    hidden: ['deletes'],
    empty: mysqlConfig,
    validate,
    merge: (index: number, form: Form) => {
      const c = deepCopy($profile.config)
      c.snapshot!.mysql![index] = form as MySQLConfig
      return c
    },
  }

  const nvidiaApp = {
    name: 'Nvidia',
    id: page.id + '.Nvidia',
    logo: nvidiaLogo,
    hidden: ['deletes'],
    merge: (index: number, form: Form) => {
      const c = deepCopy($profile.config)
      c.snapshot!.nvidia = form as NvidiaConfig
      return c
    },
  }

  async function submit() {
    const c = { ...$profile.config }
    c.snapshot!.mysql = mysql
    c.snapshot!.nvidia = nvidia
    await profile.writeConfig(c)
    m?.clear() // clears the delete counter.
  }

  let removed = $state<number[]>([])

  const allValid = $derived(
    Object.values(invalid).every(v => Object.values(v).every(v => !v)),
  )
</script>

<Header {page} />

<CardBody class="pt-0 mt-0">
  <Instances
    {validate}
    remove={index => removed.push(index)}
    bind:instances={mysql}
    bind:this={m}
    original={originalMysql}
    app={mysqlApp} />
  <InstanceHeader app={nvidiaApp} changed={!deepEqual(nvidia, originalNvidia)} />
  <Instance bind:form={nvidia!} original={originalNvidia!} app={nvidiaApp} />
</CardBody>

<Footer {submit} saveDisabled={(!formChanged && removed.length === 0) || !allValid} />
