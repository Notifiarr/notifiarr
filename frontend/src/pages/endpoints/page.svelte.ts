import type { App } from '../../includes/formsTracker.svelte'
import type { Config, Endpoint } from '../../api/notifiarrConfig'
import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { profile } from '../../api/profile.svelte'
import { deepCopy } from '../../includes/util'
import { faWebhook } from '@fortawesome/sharp-duotone-solid-svg-icons'

export const page = {
  id: 'Endpoints',
  i: faWebhook,
  c1: 'sandybrown',
  c2: 'saddlebrown',
  d1: 'lightcoral',
  d2: 'tan',
}

const empty: Endpoint = {
  name: '',
  url: '',
  method: 'POST',
  body: '',
  template: '',
  follow: false,
  frequency: 0,
  interval: 0,
  validSsl: true,
}

const merge = (index: number, form: Endpoint): Config => {
  const c = deepCopy(get(profile).config)
  if (!c.endpoints) c.endpoints = []
  for (let i = 0; i < c.endpoints.length; i++) {
    if (i === index) c.endpoints[i] = form
    else c.endpoints[i] = {} as Endpoint
  }
  return c
}

const validator = (id: string, value: any): string => {
  id = id.split('.').pop() ?? id
  if (id === 'path') {
    if (!value) return get(_)('FileWatcher.path.required')
  } else if (id === 'regex') {
    if (!value) return get(_)('FileWatcher.regex.required')
  }
  return ''
}

export const app: App<Endpoint> = {
  name: 'Endpoints',
  id: 'Endpoints',
  logo: faWebhook,
  iconProps: { c1: 'cyan', c2: 'violet', d1: 'violet', d2: 'cyan' },
  disabled: [],
  hidden: [],
  empty,
  merge,
  validator,
}
