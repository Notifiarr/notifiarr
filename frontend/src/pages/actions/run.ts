import { get } from 'svelte/store'
import { getUi, type BackendResponse } from '../../api/fetch'
import { type CronJob, type TriggerInfo } from '../../api/notifiarrConfig'
import { profile } from '../../api/profile.svelte'
import { reload } from '../../header/Reload.svelte'
import { age, warning } from '../../includes/util'
import { _ } from '../../includes/Translate.svelte'
import { cronDesc } from '../endpoints/schedule'

const reloadClient = async () => {
  try {
    await reload()
    await profile.refresh()
  } catch (error) {
    warning(`${error}`)
  }
}

// Some triggers require a specific input to be passed. This function returns that input value.
export const option = (info: TriggerInfo): string => {
  if (info.key == 'TrigCustomCommand')
    // This one requires a hash of the command.
    return get(profile).config.commands?.find(c => c.name == info.name)?.hash || 'rip'

  if (info.key == 'TrigCustomCronTimer')
    // This one requires the index of the cron.
    return `${get(profile).siteCrons?.findIndex(cron => info.name.endsWith(`'${cron.name}'`))}`

  if (info.key == 'TrigEndpointURL')
    // This one requires the name of the endpoint.
    return info.name

  return ''
}

export const run = async (info: TriggerInfo, content?: any): Promise<BackendResponse> => {
  const now = Date.now()
  if (info.key == 'TrigStop') {
    // Special case for the reload client trigger.
    await reloadClient()
    return { ok: true, body: '' }
  }

  content = encodeURIComponent(content ?? option(info))
  const url = ['trigger', info.key, content].filter(v => v).join('/')

  const resp = await getUi(url + '?ts=' + now, false)
  if (resp.ok) await profile.refresh()
  else return resp

  const diff = Date.now() - now
  // It always takes at least 1 second to run.
  if (diff < 1000) await new Promise(resolve => setTimeout(resolve, 1000 - diff))
  return resp
}

/** Formats the interval or schedule. */
export const dur = (row: TriggerInfo): string => {
  if (row.kind === 'Trigger') return get(_)(`Actions.titles.Never`)
  if (row.kind === 'Timer')
    return get(_)(`phrases.EveryDuration`, {
      values: { timeDuration: age(row.interval ?? 0, true) },
    })
  return `${cronDesc(row.cron ?? ({} as CronJob))}`
}

/** Formats and translates the name of the action, used for sorting and filtering. */
export const val = (row: TriggerInfo): string => {
  let name = row.key == 'TrigCustomCronTimer' ? row.name.split("'")[1] : row.name
  return get(_)(`Actions.triggers.${row.key}.label`, { values: { name } })
}
