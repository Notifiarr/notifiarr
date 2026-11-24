import type { App } from '../../includes/formsTracker.svelte'
import type { Config } from '../../api/notifiarrConfig'
import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { profile } from '../../api/profile.svelte'
import type { WatchFile } from '../../api/notifiarrConfig'
import { deepCopy } from '../../includes/util'
import Eye from 'phosphor-svelte/lib/Eye'

export const page = { id: 'FileWatcher' }

const empty: WatchFile = {
  path: '',
  regex: '',
  skip: '',
  poll: false,
  pipe: false,
  mustExist: false,
  logMatch: false,
  disabled: false,
}

const merge = (index: number, form: WatchFile): Config => {
  const c = deepCopy(get(profile).config)
  if (!c.watchFiles) c.watchFiles = []
  for (let i = 0; i < c.watchFiles.length; i++) {
    if (i === index) c.watchFiles[i] = form
    else c.watchFiles[i] = {} as WatchFile
  }
  return c
}

const normalizePath = (path: string, windows: boolean): string => {
  let n = path.trim().replaceAll('\\', '/')
  if (windows) n = n.toLowerCase()

  const unc = n.startsWith('//')
  const rooted = unc || n.startsWith('/') || (windows && /^[a-z]:\//.test(n))
  const stack: string[] = []
  for (const part of n.split('/')) {
    if (part === '' || part === '.') continue
    if (part === '..') {
      const head = stack.at(-1)
      if (head && head !== '..' && !(windows && /^[a-z]:$/.test(head))) stack.pop()
      else if (!rooted) stack.push('..')
      continue
    }
    stack.push(part)
  }

  let out = stack.join('/')
  if (unc) out = '//' + out
  else if (n.startsWith('/')) out = '/' + out
  return out.replace(/\/+$/, '')
}

const isClientLogPath = (path: string): boolean => {
  const p = get(profile)
  const windows = !!p.isWindows
  const want = normalizePath(path, windows)
  if (!want) return false

  const candidates = [
    p.config?.logFile,
    p.config?.httpLog,
    p.config?.debugLog,
    p.config?.services?.logFile,
    ...(p.logFileInfo?.list ?? [])
      .filter((f): f is NonNullable<typeof f> => !!f?.used)
      .map(f => f.path),
  ]

  return candidates.some(
    c => typeof c === 'string' && c !== '' && normalizePath(c, windows) === want,
  )
}

const validator = (id: string, value: any): string => {
  id = id.split('.').pop() ?? id
  if (id === 'path') {
    if (!value) return get(_)('FileWatcher.path.required')
    if (isClientLogPath(value)) return get(_)('FileWatcher.path.clientLog')
  } else if (id === 'regex') {
    if (!value) return get(_)('FileWatcher.regex.required')
  }
  return ''
}

export const app: App<WatchFile> = {
  id: 'FileWatcher',
  name: 'FileWatcher',
  envPrefix: 'WATCH_FILE',
  logo: Eye,
  iconProps: { c1: 'steelblue', d1: 'lightsteelblue' },
  disabled: [],
  hidden: [],
  empty,
  merge,
  validator,
}
