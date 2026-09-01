#!/usr/bin/env bash
# Run svelte-check and emit GitHub Actions annotations (Files / job summary).
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$root/frontend"

warn_ignore='options_missing_custom_element:ignore'
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

set +e
npx svelte-check --tsconfig ./tsconfig.app.json --compiler-warnings "$warn_ignore" --output machine >"$tmp"
sc=$?
set -e

export SVELTE_CHECK_MACHINE="$tmp"
node --input-type=module <<'NODE'
import fs from 'node:fs'

const text = fs.readFileSync(process.env.SVELTE_CHECK_MACHINE, 'utf8')
const re = /^(ERROR|WARNING) "([^"]+)" (\d+):(\d+) "([\s\S]*?)"$/gm
let errors = 0
let warnings = 0
let m
while ((m = re.exec(text))) {
  const kind = m[1]
  const file = m[2]
  const line = m[3]
  const col = m[4]
  const msg = m[5].replace(/\n/g, ' ').replace(/\\n/g, ' ')
  const sev = kind === 'WARNING' ? 'warning' : 'error'
  if (sev === 'warning') warnings++
  else errors++
  console.log(`::${sev} file=frontend/${file},line=${line},col=${col}::${msg}`)
  console.log(`${kind} frontend/${file}:${line}:${col}`)
  console.log(`  ${msg}`)
}
const completed = text.split('\n').find(l => l.includes(' COMPLETED '))
if (completed) console.log(completed)
console.log(`svelte-check found ${errors} errors and ${warnings} warnings`)
if (process.env.GITHUB_STEP_SUMMARY) {
  const extra =
    errors > 0 ? '\nErrors are also listed as annotations on the **Files** tab.\n' : ''
  fs.appendFileSync(
    process.env.GITHUB_STEP_SUMMARY,
    `## svelte-check\n\n| Errors | Warnings |\n| ---: | ---: |\n| ${errors} | ${warnings} |\n${extra}`,
  )
}
NODE

npx tsc -p tsconfig.node.json
tsc=$?

if [[ "$sc" -ne 0 ]]; then
  exit "$sc"
fi
exit "$tsc"
