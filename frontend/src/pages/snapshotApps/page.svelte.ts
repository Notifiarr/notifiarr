import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { deepCopy } from '../../includes/util'
import type { MySQLConfig, NvidiaConfig } from '../../api/notifiarrConfig'
import { profile } from '../../api/profile.svelte'
import mysqlLogo from '../../assets/logos/mysql.png'
import nvidiaLogo from '../../assets/logos/nvidia.png'
import { type App } from '../../includes/formsTracker.svelte'
import { validate } from '../../includes/instanceValidator'

export const page = { id: 'SnapshotApps' }

export class SnapshotApps {
  static get title(): Record<string, string> {
    return {
      ['MySQL']: get(_)('SnapshotApps.MySQL.title'),
      ['Nvidia']: get(_)('SnapshotApps.Nvidia.title'),
    }
  }

  static readonly mysqlApp: App<MySQLConfig> = {
    name: 'MySQL',
    id: page.id + '.MySQL',
    envPrefix: 'SNAPSHOT_MYSQL',
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
        return value === '' ? get(_)('phrases.UsernameMustNotBeEmpty') : ''
      return validate(id, value, index, instances)
    },
    merge: (index: number, form: MySQLConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.snapshot) c.snapshot = { mysql: [] } as typeof c.snapshot
      c.snapshot.mysql ??= []
      c.snapshot.mysql[index] = form
      return c
    },
  }

  static readonly nvidiaApp: App<NvidiaConfig> = {
    name: 'Nvidia',
    id: page.id + '.Nvidia',
    logo: nvidiaLogo,
    envPrefix: 'SNAPSHOT_NVIDIA',
    hidden: ['deletes'],
    empty: { busIDs: [''], smiPath: '', disabled: false },
    merge: (index: number, form: NvidiaConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.snapshot) c.snapshot = { nvidia: form } as typeof c.snapshot
      else c.snapshot.nvidia = form
      return c
    },
  }

  // Keep track of the navigation.
  static readonly tabs = ['mysql', 'nvidia']
}
