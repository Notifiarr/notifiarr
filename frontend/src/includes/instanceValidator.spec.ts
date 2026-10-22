import { describe, expect, it } from 'vitest'
import type { Form } from './Instance.svelte'
import { validate } from './instanceValidator'

const forms = (...names: string[]): Form[] => names.map(name => ({ name }))

describe('validate', () => {
  it('rejects an empty name', () => {
    expect(validate('StarrApps.Sonarr.name', '', 0, forms(''))).toBe(
      'Name must not be empty.',
    )
  })

  it('rejects a name another instance already uses', () => {
    expect(validate('StarrApps.Sonarr.name', 'Main', 0, forms('Main', 'Main'))).toBe(
      'This name is in use by instance 2.',
    )
  })

  it('accepts a name that matches only the instance being edited', () => {
    expect(validate('StarrApps.Sonarr.name', 'Main', 0, forms('Main', 'Other'))).toBe('')
  })

  it('requires an http or https URL', () => {
    expect(validate('apps.url', 'ftp://nas', 0, [])).toBe(
      'URL must begin with http:// or https://',
    )
    expect(validate('apps.url', undefined, 0, [])).toBe(
      'URL must begin with http:// or https://',
    )
    expect(validate('apps.url', 'http://nas', 0, [])).toBe('')
    expect(validate('apps.url', 'https://nas', 0, [])).toBe('')
  })

  it('rejects an empty host and accepts a filled one', () => {
    expect(validate('apps.host', '', 0, [])).toBe('Host must not be empty.')
    expect(validate('apps.host', 'nas.local', 0, [])).toBe('')
  })

  it('requires an api key of at least 32 characters', () => {
    expect(validate('apps.apiKey', 'short', 0, [])).toBe(
      'API key must be 32 characters or longer.',
    )
    expect(validate('apps.apiKey', 'a'.repeat(32), 0, [])).toBe('')
  })

  it('requires a token of at least 8 characters', () => {
    expect(validate('apps.token', 'short', 0, [])).toBe(
      'Token must be 8 characters or longer.',
    )
    expect(validate('apps.token', '12345678', 0, [])).toBe('')
  })

  it('ignores fields it does not validate', () => {
    expect(validate('apps.timeout', '10s', 0, [])).toBe('')
  })
})
