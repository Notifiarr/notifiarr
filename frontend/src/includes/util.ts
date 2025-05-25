import { toast } from '@zerodevx/svelte-toast'
import { get } from 'svelte/store'
import { _ } from './Translate.svelte'

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

/** Convert a date into a human readable string of how long ago it was. */
export function since(date: Date | string): string {
  const now = new Date().getTime()
  const then = typeof date === 'string' ? new Date(date).getTime() : date.getTime()
  return age(now - then, false)
}

/** age converts a milliseconds counter into human readable: 13h 5m 45s */
export function age(milliseconds: number, includeSeconds = false): string {
  const seconds = Math.floor(milliseconds / 1000)
  if (!seconds) return '0s'
  const weeks = Math.floor(seconds / 604800)
  const days = Math.floor((seconds - weeks * 604800) / 86400)
  const hours = Math.floor((seconds - weeks * 604800 - days * 86400) / 3600)
  const minutes = Math.floor(
    (seconds - weeks * 604800 - days * 86400 - hours * 3600) / 60,
  )
  const secs =
    !includeSeconds && seconds > 60
      ? 0
      : Math.floor(seconds - weeks * 604800 - days * 86400 - hours * 3600 - minutes * 60)
  return (
    (weeks > 0 ? get(_)('words.clock.short.w', { values: { weeks } }) + ' ' : '') +
    (days > 0 ? get(_)('words.clock.short.d', { values: { days } }) + ' ' : '') +
    (hours > 0 ? get(_)('words.clock.short.h', { values: { hours } }) + ' ' : '') +
    (minutes > 0 ? get(_)('words.clock.short.m', { values: { minutes } }) + ' ' : '') +
    (seconds > 0 ? get(_)('words.clock.short.s', { values: { seconds } }) + ' ' : '')
  ).trim()
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
    values: { bytes: bytes.toFixed(2) },
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

/** Deep copy an object. */
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
