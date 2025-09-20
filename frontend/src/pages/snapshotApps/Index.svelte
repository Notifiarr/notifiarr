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
  import { deepCopy } from '../../includes/util'
  import Instance from '../../includes/Instance.svelte'
  import InstanceHeader from '../../includes/InstanceHeader.svelte'
  import { FormListTracker, type App } from '../../includes/formsTracker.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import { validate } from '../../includes/instanceValidator'

  const mysqlApp: App<MySQLConfig> = {
    name: 'MySQL',
    id: page.id + '.MySQL',
    logo: mysqlLogo,
    hidden: ['deletes'],
    empty: {
      name: '',
      host: '',
      username: '',
      password: '',
      timeout: '10s',
      interval: '5m0s',
    },
    validator: (id: string, value: any, index: number, instances: MySQLConfig[]) => {
      if (id.endsWith('.username'))
        return value === '' ? $_('phrases.UsernameMustNotBeEmpty') : ''
      return validate(id, value, index, instances)
    },
    merge: (index: number, form: MySQLConfig) => {
      const c = deepCopy($profile.config)
      c.snapshot.mysql![index] = form
      return c
    },
  }

  const nvidiaApp: App<NvidiaConfig> = {
    name: 'Nvidia',
    id: page.id + '.Nvidia',
    logo: nvidiaLogo,
    hidden: ['deletes'],
    empty: { busIDs: [''], smiPath: '', disabled: false },
    merge: (index: number, form: NvidiaConfig) => {
      const c = deepCopy($profile.config)
      c.snapshot.nvidia = form
      return c
    },
  }

  if (!$profile.config.snapshot.nvidia.busIDs?.length) {
    $profile.config.snapshot.nvidia.busIDs = ['']
  }

  let flt = $derived({
    MySQL: new FormListTracker($profile.config.snapshot.mysql ?? [], mysqlApp),
    Nvidia: new FormListTracker([$profile.config.snapshot.nvidia], nvidiaApp),
  })

  async function submit() {
    const c = { ...$profile.config }
    c.snapshot.mysql = flt.MySQL.instances as MySQLConfig[]
    c.snapshot.nvidia = flt.Nvidia.instances[0]
    await profile.writeConfig(c)
  }

  $effect(() => {
    nav.formChanged = Object.values(flt).some(iv => iv.formChanged)
  })
</script>

<Header {page} />

<CardBody class="pt-0 mt-0">
  <Instances flt={flt.MySQL} Child={Instance}>
    {#snippet headerActive(index)}
      {index + 1}. {flt.MySQL.original?.[index]?.name}
    {/snippet}
    {#snippet headerCollapsed(index)}
      {flt.MySQL.original?.[index]?.host}
    {/snippet}
  </Instances>
  <InstanceHeader flt={flt.Nvidia} />
  <Instance
    index={0}
    bind:form={flt.Nvidia.instances[0]}
    original={flt.Nvidia.original[0]}
    app={nvidiaApp} />
</CardBody>

<Footer
  {submit}
  saveDisabled={!nav.formChanged || Object.values(flt).some(iv => iv.invalid)} />
