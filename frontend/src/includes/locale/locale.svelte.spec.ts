import { get } from 'svelte/store'
import { _, date, locale as svelteLocale } from 'svelte-i18n'
import { describe, expect, it } from 'vitest'
import { locale } from './locale.svelte'

describe('Traditional Chinese locale', () => {
  it('formats counts and dates with a BCP-47 tag', async () => {
    await locale.set('zh-Hant')

    expect(locale.current).toBe('zh-Hant')
    expect(get(svelteLocale)).toBe('zh-Hant')
    expect(get(_)('ProcessList.footer', { values: { count: 42 } })).toBe(
      '找到 42 個程序。',
    )
    expect(get(date)(new Date('2026-01-02T00:00:00Z'))).toEqual(expect.any(String))
  })

  it('keeps a plain language code and collapses a regional tag', async () => {
    await locale.set('en-US')

    expect(locale.current).toBe('en')
    expect(get(svelteLocale)).toBe('en')
    expect(get(_)('ProcessList.footer', { values: { count: 42 } })).toBe(
      '42 processes found.',
    )
  })
})

describe('Simplified Chinese locale', () => {
  it('collapses a zh regional tag to zh-Hans', async () => {
    await locale.set('zh-CN')

    expect(locale.current).toBe('zh-Hans')
    expect(get(svelteLocale)).toBe('zh-Hans')
    expect(get(_)('ProcessList.footer', { values: { count: 42 } })).toBe(
      '找到 42 个进程。',
    )
  })

  it('maps a bare zh tag to zh-Hans', async () => {
    await locale.set('zh')

    expect(locale.current).toBe('zh-Hans')
    expect(get(svelteLocale)).toBe('zh-Hans')
  })

  it('keeps Hong Kong and Taiwan tags on zh-Hant', async () => {
    await locale.set('zh-HK')
    expect(locale.current).toBe('zh-Hant')

    await locale.set('zh-TW')
    expect(locale.current).toBe('zh-Hant')
  })
})
