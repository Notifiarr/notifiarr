import type { App } from '../../includes/formsTracker.svelte'
import type { Config } from '../../api/notifiarrConfig'
import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { profile } from '../../api/profile.svelte'
import type { WatchFile } from '../../api/notifiarrConfig'
import { deepCopy } from '../../includes/util'
import { faEyeEvil, faFileWaveform } from '@fortawesome/sharp-duotone-light-svg-icons'

export const page = {
  id: 'FileWatcher',
  i: faFileWaveform,
  d1: 'thistle',
  d2: 'blue',
  c1: 'sienna',
  c2: 'moccasin',
}

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
  let n = path.trim().replaceAll('\\', '/').replace(/\/+$/, '')
  if (windows) n = n.toLowerCase()
  return n
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
    ...(p.logFileInfo?.list ?? []).filter(f => f.used).map(f => f.path),
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
  logo: faEyeEvil,
  iconProps: { c1: 'purple', c2: 'firebrick', d1: 'thistle', d2: 'violet' },
  disabled: [],
  hidden: [],
  empty,
  merge,
  validator,
}
