import { toast } from '@zerodevx/svelte-toast'

/** Remove a prefix from a string. */
export const ltrim = (str: string, prefix: string) =>
  str.slice(str.startsWith(prefix) ? prefix.length : 0)

/** Remove a suffix from a string. */
export const rtrim = (str: string, suffix: string) =>
  str.endsWith(suffix) ? str.slice(0, str.length - suffix.length) : str

/** Show a success toast. */
export const success = (m: string) =>
  toast.push(m, {
    theme: {
      '--toastBackground': 'green',
      '--toastColor': 'white',
      '--toastBarBackground': 'olive',
    },
  })

/** Show a warning toast. */
export const warning = (m: string) =>
  toast.push(m, {
    theme: {
      '--toastBackground': 'orange',
      '--toastColor': 'white',
      '--toastBarBackground': 'black',
    },
  })

/** Show a failure toast. */
export const failure = (m: string) =>
  toast.push(m, {
    theme: {
      '--toastBackground': 'red',
      '--toastColor': 'white',
      '--toastBarBackground': 'royalblue',
    },
  })

/** age converts a milliseconds counter into human readable: 13h 5m 45s */
export function age(milliseconds: number, includeSeconds = false): string {
  const seconds = Math.floor(milliseconds / 1000)
  if (!seconds) return '0s'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds - days * 86400) / 3600)
  const minutes = Math.floor((seconds - days * 86400 - hours * 3600) / 60)
  const secs =
    !includeSeconds && seconds > 60
      ? 0
      : Math.floor(seconds - days * 86400 - hours * 3600 - minutes * 60)
  return (
    (days > 0 ? days + 'd ' : '') +
    (hours > 0 ? hours + 'h ' : '') +
    (minutes > 0 ? minutes + 'm ' : '') +
    (secs > 0 ? secs + 's ' : '')
  ).trim()
}

/** Add a delay anywhere in any async function. */
export const delay = (ms: number): Promise<void> =>
  new Promise(resolve => setTimeout(resolve, ms))

/** Check if two strings are equal, case insensitive. */
export const iequals = (a: string, b: string) => a.toLowerCase() === b.toLowerCase()
