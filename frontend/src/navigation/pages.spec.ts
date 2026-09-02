import { describe, expect, it } from 'vitest'
import src from './pages.ts?raw'

describe('API docs page registry', () => {
  it('does not statically import the RapiDoc page into the boot graph', () => {
    expect(src).toMatch(/id: 'ApiDocs'|ApiDocsPage/)
    expect(src).not.toMatch(/stubs\/ApiDocs\.svelte/)
    expect(src).toMatch(/stubs\/ApiDocsLazy\.svelte/)
  })
})
