import { describe, expect, it } from 'vitest'
import { formatAnnotation, parseMachine, repoFile } from './svelte-check.mjs'

/** Official svelte-check --output machine examples (timestamp prefix required). */
const readme = [
  '1590680325583 START "/home/user/language-tools/packages/language-server/test/plugins/typescript/testfiles"',
  '1590680326283 ERROR "codeactions.svelte" 1:16 "Cannot find module \'blubb\' or its corresponding type declarations."',
  '1590680326778 WARNING "imported-file.svelte" 0:37 "Component has unused export property \'prop\'."',
  '1590680326807 COMPLETED 20 FILES 21 ERRORS 1 WARNINGS 3 FILES_WITH_PROBLEMS',
].join('\n')

const failureOnly = '1590680328921 FAILURE "Connection closed"\n'

const clean = [
  '1788312682085 START "/Users/david/.cursor/worktrees/phosphor-chrome/notifiarr/frontend"',
  '1788312682095 COMPLETED 3601 FILES 0 ERRORS 0 WARNINGS 0 FILES_WITH_PROBLEMS',
].join('\n')

const oldRe = /^(ERROR|WARNING) "([^"]+)" (\d+):(\d+) "([\s\S]*?)"$/gm

describe('svelte-check machine parser', () => {
  it('does not match timestamped records without the prefix (the old regex)', () => {
    expect([...readme.matchAll(oldRe)]).toHaveLength(0)
  })

  it('parses ERROR and WARNING records from the svelte-check docs', () => {
    const { diagnostics, completed, failure } = parseMachine(readme)
    expect(failure).toBeNull()
    expect(diagnostics).toEqual([
      {
        kind: 'ERROR',
        file: 'codeactions.svelte',
        line: 1,
        col: 16,
        msg: "Cannot find module 'blubb' or its corresponding type declarations.",
      },
      {
        kind: 'WARNING',
        file: 'imported-file.svelte',
        line: 0,
        col: 37,
        msg: "Component has unused export property 'prop'.",
      },
    ])
    expect(completed).toEqual({
      files: 20,
      errors: 21,
      warnings: 1,
      filesWithProblems: 3,
      text: 'COMPLETED 20 FILES 21 ERRORS 1 WARNINGS 3 FILES_WITH_PROBLEMS',
    })
  })

  it('parses a FAILURE record', () => {
    const { diagnostics, completed, failure } = parseMachine(failureOnly)
    expect(diagnostics).toEqual([])
    expect(completed).toBeNull()
    expect(failure).toBe('Connection closed')
  })

  it('parses a clean COMPLETED run', () => {
    const { diagnostics, completed, failure } = parseMachine(clean)
    expect(diagnostics).toEqual([])
    expect(failure).toBeNull()
    expect(completed?.errors).toBe(0)
    expect(completed?.warnings).toBe(0)
    expect(completed?.files).toBe(3601)
  })

  it('builds a repo-relative GitHub annotation', () => {
    const { diagnostics } = parseMachine(readme)
    expect(repoFile(diagnostics[0].file)).toBe('frontend/codeactions.svelte')
    expect(formatAnnotation(diagnostics[0])).toBe(
      "::error file=frontend/codeactions.svelte,line=1,col=16::Cannot find module 'blubb' or its corresponding type declarations.",
    )
    expect(formatAnnotation(diagnostics[1])).toBe(
      "::warning file=frontend/imported-file.svelte,line=0,col=37::Component has unused export property 'prop'.",
    )
  })

  it('JSON-unescapes quoted file and message fields', () => {
    const file = 'node_modules/@zerodevx/svelte-toast/dist/stores.d.ts'
    const msg = `Namespace "svelte/store" has no exported member 'Invalidator'.`
    const line = `1590680326283 ERROR ${JSON.stringify(file)} 91:122 ${JSON.stringify(msg)}`
    const { diagnostics } = parseMachine(line + '\n')
    expect(diagnostics).toEqual([{ kind: 'ERROR', file, line: 91, col: 122, msg }])
    const ann = formatAnnotation(diagnostics[0])
    expect(ann).toContain('Namespace "svelte/store"')
    expect(ann).not.toContain('\\"')
  })

  it('turns escaped newlines into GitHub %0A', () => {
    const msg = 'first\nsecond'
    const line = `1 ERROR ${JSON.stringify('a.svelte')} 1:1 ${JSON.stringify(msg)}`
    const { diagnostics } = parseMachine(line)
    expect(diagnostics[0].msg).toBe(msg)
    expect(formatAnnotation(diagnostics[0])).toContain('first%0Asecond')
  })
})
