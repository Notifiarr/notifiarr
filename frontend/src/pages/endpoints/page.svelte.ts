import type { App } from '../../includes/formsTracker.svelte'
import type { Config, Endpoint } from '../../api/notifiarrConfig'
import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { profile } from '../../api/profile.svelte'
import { deepCopy } from '../../includes/util'
import { faLinkSimple, faWebhook } from '@fortawesome/sharp-duotone-solid-svg-icons'
import { validator as cronValidator } from './CronScheduler.svelte'

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
  method: 'GET',
  body: '',
  template: '',
  follow: false,
  frequency: 0,
  interval: 0,
  timeout: '0s',
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

const validator = (
  id: string,
  value: any,
  index: number,
  instances: Endpoint[],
): string => {
  id = id.split('.').pop() ?? id

  if (id == 'name') {
    let found = ''
    instances?.forEach((m, i) => {
      if (i !== index && m?.name === value) {
        found = get(_)('phrases.NameInUseByInstance', { values: { number: i + 1 } })
        return
      }
    })
    if (found) return found
    return value ? '' : get(_)('phrases.NameMustNotBeEmpty')
  } else if (id == 'url') {
    return value.match(/^http:\/\/../) || value.match(/^https:\/\/../)
      ? ''
      : get(_)('phrases.URLMustBeginWithHttp')
  } else if (id === 'template') {
    return value ? '' : get(_)('Endpoints.template.required')
  }

  return cronValidator(id, value)
}

export const app: App<Endpoint> = {
  name: 'Endpoints',
  id: 'Endpoints',
  logo: faLinkSimple,
  iconProps: { c1: 'orange', c2: 'violet' },
  disabled: [],
  hidden: [],
  empty,
  merge,
  validator,
}
