#!/usr/bin/env node
/** Run svelte-check (machine output) and tsc; emit GitHub Actions annotations. */

import { spawnSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const frontend = path.resolve(import.meta.dirname, '..')
const warnIgnore = 'options_missing_custom_element:ignore'

const quoted = '"(?:[^"\\\\]|\\\\.)*"'
const diagRe = new RegExp(
  `^(\\d+) (ERROR|WARNING) (${quoted}) (\\d+):(\\d+) (${quoted})$`,
)
const completedRe =
  /^(\d+) COMPLETED (\d+) FILES (\d+) ERRORS (\d+) WARNINGS(?: (\d+) FILES_WITH_PROBLEMS)?/
const failureRe = new RegExp(`^(\\d+) FAILURE (${quoted})$`)
const maxBuffer = 256 * 1024 * 1024

export function parseMachine(text) {
  const diagnostics = []
  let completed = null
  let failure = null
  for (const line of text.split(/\r?\n/)) {
    if (!line) continue
    let m = line.match(diagRe)
    if (m) {
      diagnostics.push({
        kind: m[2],
        file: JSON.parse(m[3]),
        line: Number(m[4]),
        col: Number(m[5]),
        msg: JSON.parse(m[6]),
      })
      continue
    }
    m = line.match(completedRe)
    if (m) {
      completed = {
        files: Number(m[2]),
        errors: Number(m[3]),
        warnings: Number(m[4]),
        filesWithProblems: m[5] != null ? Number(m[5]) : 0,
        text: line.replace(/^\d+ /, ''),
      }
      continue
    }
    m = line.match(failureRe)
    if (m) failure = JSON.parse(m[2])
  }
  return { diagnostics, completed, failure }
}

export function repoFile(file) {
  if (path.isAbsolute(file)) return file
  return path.posix.join('frontend', file.replaceAll('\\', '/'))
}

function ghaEscape(s) {
  return s.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A')
}

export function formatAnnotation(diag) {
  const sev = diag.kind === 'WARNING' ? 'warning' : 'error'
  return `::${sev} file=${repoFile(diag.file)},line=${diag.line},col=${diag.col}::${ghaEscape(diag.msg)}`
}

function run(bin, args) {
  return spawnSync(process.execPath, [bin, ...args], {
    cwd: frontend,
    encoding: 'utf8',
    maxBuffer,
  })
}

function report({ diagnostics, completed, failure }) {
  const gha = process.env.GITHUB_ACTIONS === 'true'
  for (const d of diagnostics) {
    if (gha) console.log(formatAnnotation(d))
    console.log(`${d.kind} ${repoFile(d.file)}:${d.line}:${d.col}`)
    console.log(`  ${d.msg}`)
  }
  if (failure) {
    if (gha) console.log(`::error::${ghaEscape(failure)}`)
    console.log(`FAILURE ${failure}`)
  }
  if (completed) console.log(completed.text)
  const errors = diagnostics.filter(d => d.kind === 'ERROR').length
  const warnings = diagnostics.filter(d => d.kind === 'WARNING').length
  console.log(`svelte-check found ${errors} errors and ${warnings} warnings`)
  if (process.env.GITHUB_STEP_SUMMARY) {
    const extra =
      errors > 0 ? '\nErrors are also listed as annotations on the **Files** tab.\n' : ''
    const failLine = failure ? `\nChecker failure: ${failure}\n` : ''
    fs.appendFileSync(
      process.env.GITHUB_STEP_SUMMARY,
      `## svelte-check\n\n| Errors | Warnings |\n| ---: | ---: |\n| ${errors} | ${warnings} |\n${extra}${failLine}`,
    )
  }
  if (completed && (completed.errors !== errors || completed.warnings !== warnings)) {
    console.error(
      `parser mismatch: counted ${errors}/${warnings} but COMPLETED says ${completed.errors}/${completed.warnings}`,
    )
    return 1
  }
  return 0
}

function main() {
  const svelteCheck = path.join(frontend, 'node_modules/svelte-check/bin/svelte-check')
  const tsc = path.join(frontend, 'node_modules/typescript/bin/tsc')
  const sc = run(svelteCheck, [
    '--tsconfig',
    './tsconfig.app.json',
    '--compiler-warnings',
    warnIgnore,
    '--output',
    'machine',
  ])
  if (sc.error) {
    console.error(sc.error)
    return 1
  }
  if (sc.stderr) process.stderr.write(sc.stderr)
  const parsed = parseMachine(sc.stdout ?? '')
  const parseExit = report(parsed)
  const ts = run(tsc, ['-p', 'tsconfig.node.json'])
  if (ts.error) {
    console.error(ts.error)
    return 1
  }
  if (ts.stdout) process.stdout.write(ts.stdout)
  if (ts.stderr) process.stderr.write(ts.stderr)
  if (parsed.failure) return sc.status || 1
  if (parseExit) return parseExit
  if (sc.status) return sc.status
  return ts.status ?? 1
}

const thisFile = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === thisFile) {
  process.exit(main())
}
