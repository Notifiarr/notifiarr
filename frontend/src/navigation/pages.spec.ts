import { describe, expect, it } from 'vitest'
import src from './pages.ts?raw'
import lazySrc from '../pages/stubs/ApiDocsLazy.svelte?raw'

describe('API docs page registry', () => {
  it('does not statically import the RapiDoc page into the boot graph', () => {
    expect(src).toMatch(/id: 'ApiDocs'|ApiDocsPage/)
    expect(src).not.toMatch(/stubs\/ApiDocs\.svelte/)
    expect(src).toMatch(/stubs\/ApiDocsLazy\.svelte/)
  })

  it('waits for the page slide before mounting RapiDoc', () => {
    expect(lazySrc).toMatch(/afterPageSlide/)
    expect(lazySrc).toMatch(/await tick/)
    expect(lazySrc).not.toMatch(/^\s*const loaded = import\('\.\/ApiDocs\.svelte'\)/m)
  })
})
