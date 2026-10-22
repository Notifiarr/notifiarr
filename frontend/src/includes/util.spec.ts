import { describe, expect, it } from 'vitest'
import { age, deepCopy, deepEqual, escapeHtml, formatBytes, present } from './util'

describe('deepEqual', () => {
  it('matches nested values and rejects a different shape', () => {
    expect(deepEqual({ a: { b: 1 }, c: [2] }, { a: { b: 1 }, c: [2] })).toBe(true)
    expect(deepEqual({ a: 1 }, { a: 1, b: 2 })).toBe(false)
    expect(deepEqual({ a: { b: 1 } }, { a: { b: 2 } })).toBe(false)
    expect(deepEqual(null, null)).toBe(true)
  })
})

describe('deepCopy', () => {
  it('copies nested objects without sharing them', () => {
    const original = { name: 'sonarr', days: [1, 2], nested: { on: true } }
    const copy = deepCopy(original)

    copy.name = 'radarr'
    copy.days.push(3)
    copy.nested.on = false

    expect(original).toEqual({ name: 'sonarr', days: [1, 2], nested: { on: true } })
    expect(copy).toEqual({ name: 'radarr', days: [1, 2, 3], nested: { on: false } })
  })

  it('returns primitives and null unchanged', () => {
    expect(deepCopy('ok')).toBe('ok')
    expect(deepCopy(null)).toBe(null)
  })
})

describe('formatBytes', () => {
  it('uses the translated unit for bytes, kilobytes, and megabytes', () => {
    expect(formatBytes(0)).toBe('0 B')
    expect(formatBytes(512)).toBe('512 B')
    expect(formatBytes(1024)).toBe('1.0 KB')
    expect(formatBytes(1536)).toBe('1.5 KB')
    expect(formatBytes(5 * 1024 * 1024)).toBe('5.0 MB')
  })
})

describe('age', () => {
  it('formats milliseconds, seconds, and minutes', () => {
    expect(age(0)).toBe('0ms')
    expect(age(500)).toBe('500ms')
    expect(age(5000)).toBe('5s')
    expect(age(90_000)).toBe('1m')
    expect(age(90_000, true)).toBe('1m 30s')
  })

  it('keeps weeks, days, hours, and seconds when asked', () => {
    const ms = (2 * 604800 + 1 * 86400 + 3 * 3600 + 4 * 60 + 5) * 1000
    expect(age(ms, true)).toBe('2w 1d 3h 4m 5s')
  })
})

describe('present', () => {
  it('drops null holes and missing lists', () => {
    expect(present(null)).toEqual([])
    expect(present(undefined)).toEqual([])
    expect(present([1, null, 2, undefined])).toEqual([1, 2])
  })
})

describe('escapeHtml', () => {
  it('escapes the characters translations interpolate as HTML', () => {
    expect(escapeHtml(`a&b<c>d"e'f`)).toBe('a&amp;b&lt;c&gt;d&quot;e&#039;f')
  })
})
