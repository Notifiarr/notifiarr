import { get } from 'svelte/store'
import { _ } from './Translate.svelte'
import type { Form } from './Instance.svelte'

/** Standard form validator for an integrated instance (plex, sonarr, etc)
 * @param id - The id of the form field. (anything.here.url)
 * @param value - The value of the form field. (http://localhost:8080)
 * @param index - The index of the current instance the instances list. (0)
 * @param instances - The instances list to verify unique names against.
 * @returns The feedback for the instance.
 */
export const validate = (
  id: string,
  value: any,
  index: number,
  instances: Form[],
): string => {
  const key = id.split('.').pop()

  if (key == 'name') {
    let found = ''
    instances?.forEach((m, i) => {
      if (i !== index && m?.name === value) {
        found = get(_)('phrases.NameInUseByInstance', { values: { number: i + 1 } })
        return
      }
    })
    if (found) return found
    return value ? '' : get(_)('phrases.NameMustNotBeEmpty')
  } else if (key == 'url') {
    return value.startsWith('http://') || value.startsWith('https://')
      ? ''
      : get(_)('phrases.URLMustBeginWithHttp')
  } else if (key == 'host' && value === '') {
    return get(_)('phrases.HostMustNotBeEmpty')
  } else if (key == 'apiKey' && value.length < 32) {
    return get(_)('phrases.APIKeyMustBeCountCharacters', { values: { count: 32 } })
  } else if (key == 'token' && value.length < 8) {
    return get(_)('phrases.TokenMustBeCountCharacters', { values: { count: 8 } })
  }

  return ''
}
