import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

describe('API docs page registry', () => {
  it('does not statically import the RapiDoc page into the boot graph', () => {
    const src = readFileSync(
      join(dirname(fileURLToPath(import.meta.url)), 'pages.ts'),
      'utf8',
    )
    expect(src).toMatch(/id: 'ApiDocs'|ApiDocsPage/)
    expect(src).not.toMatch(/stubs\/ApiDocs\.svelte/)
    expect(src).toMatch(/stubs\/ApiDocsLazy\.svelte/)
  })
})
