import { toast } from '@zerodevx/svelte-toast'
import { get } from 'svelte/store'
import { _ } from './Translate.svelte'

/** Remove a prefix from a string. */
export const ltrim = (str: string, prefix: string) =>
  str.slice(str.startsWith(prefix) ? prefix.length : 0)
/** Remove a suffix from a string. */
export const rtrim = (str: string, suffix: string) =>
  str.endsWith(suffix) ? str.slice(0, str.length - suffix.length) : str

/** Show a success toast if the translation key exists. */
export const successIf = (key: string) => {
  const t = get(_)(key)
  if (t != key) success(t)
}

/** Show a success toast. */
export const success = (m: string) => {
  const e = ['🎉', '🌟', '⭐️', '🏆', '🎯', '☑️', '🎊', '🎈', '🎁', '🏅', '👍', '👏']
  const r = Math.floor(Math.random() * e.length)
  toast.push(e[r] + ' &nbsp;' + m, { classes: ['toast-wrapman', 'success-toast'] })
}
// Classes for these are in assets/app.css.
/** Show a warning toast. */
export const warning = (m: string) =>
  toast.push('⚠️ &nbsp;' + m, { classes: ['toast-wrapman', 'warning-toast'] })
/** Show a failure toast. */
export const failure = (m: string) =>
  toast.push('⚠️ &nbsp;' + m, { classes: ['toast-wrapman', 'failure-toast'] })

/** Convert a date into a human readable string of how long ago it was. */
export function since(date: Date | string): string {
  const now = new Date().getTime()
  const then = typeof date === 'string' ? new Date(date).getTime() : date.getTime()
  return age(now - then, true)
}

/** age converts a milliseconds counter into human readable: 13h 5m 45s */
export function age(milliseconds: number, includeSeconds = false): string {
  const t = get(_) // translate function
  let seconds = Math.floor(milliseconds / 1000)
  if (!seconds) return t('words.clock.short.s', { values: { seconds: 0 } })
  const weeks = Math.floor(seconds / 604800)
  const days = Math.floor((seconds - weeks * 604800) / 86400)
  const hours = Math.floor((seconds - weeks * 604800 - days * 86400) / 3600)
  const minutes = Math.floor(
    (seconds - weeks * 604800 - days * 86400 - hours * 3600) / 60,
  )
  seconds =
    !includeSeconds && seconds > 60
      ? 0
      : Math.floor(seconds - weeks * 604800 - days * 86400 - hours * 3600 - minutes * 60)

  const w = weeks ? t('words.clock.short.w', { values: { weeks } }) + ' ' : ''
  const d = days ? t('words.clock.short.d', { values: { days } }) + ' ' : ''
  const h = hours ? t('words.clock.short.h', { values: { hours } }) + ' ' : ''
  const m = minutes ? t('words.clock.short.m', { values: { minutes } }) + ' ' : ''
  const s = seconds ? t('words.clock.short.s', { values: { seconds } }) + ' ' : ''

  return t('words.clock.format', {
    values: { weeks: w, days: d, hours: h, minutes: m, seconds: s, milliseconds: '' },
  }).trim()
}

/** Add a delay anywhere in any async function. */
export const delay = (ms: number): Promise<void> =>
  new Promise(resolve => setTimeout(resolve, ms))

/** Check if two strings are equal, case insensitive. */
export const iequals = (a: string, b: string) => a.toLowerCase() === b.toLowerCase()

/** Format bytes number into a human readable string. */
export function formatBytes(bytes: number): string {
  // Translation keys.
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let unitIndex = 0

  while (bytes >= 1024 && unitIndex < units.length - 1) {
    bytes /= 1024
    unitIndex++
  }

  return get(_)('words.bytes.short.' + units[unitIndex], {
    values: { bytes: bytes.toFixed(unitIndex === 0 ? 0 : 1) },
  })
}

/** Check if two objects are equal. */
export const deepEqual = (obj1: any, obj2: any): boolean => {
  if (
    typeof obj1 !== 'object' ||
    typeof obj2 !== 'object' ||
    obj1 === null ||
    obj2 === null
  ) {
    return obj1 === obj2
  }

  const keys1 = Object.keys(obj1)
  const keys2 = Object.keys(obj2)
  if (keys1.length !== keys2.length) {
    return false
  }

  for (let key of keys1) {
    if (!obj2.hasOwnProperty(key) || !deepEqual(obj1[key], obj2[key])) {
      return false
    }
  }

  return true
}

/** Deep copy an object. Works fine on our reasonably simple app config. */
export const deepCopy = <T>(obj: T): T => {
  if (typeof obj !== 'object' || obj === null) {
    return obj
  }

  if (Array.isArray(obj)) {
    return obj.map(item => deepCopy(item)) as any
  }

  const copiedObj: { [key in keyof T]?: T[key] } = {}
  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {
      copiedObj[key] = deepCopy(obj[key])
    }
  }

  return copiedObj as T
}

export const maxLength = (str: string, max: number) =>
  str.length > max ? str.slice(0, max) + ' ....' : str

export const escapeHtml = (unsafe: string) => {
  return unsafe
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;')
}

/** Count the items nested in an object, or the number of keys in an object. */
export const mapLength = (map: Record<string, any | null> | undefined): number => {
  if (!map) return 0
  let count = 0
  for (const key in map) {
    count += (map[key]?.length ?? Object.keys(map[key] ?? {}).length) || 1
  }
  return count
}
