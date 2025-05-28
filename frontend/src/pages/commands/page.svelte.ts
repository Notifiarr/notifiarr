import type { App } from '../../includes/formsTracker.svelte'
import type { Config } from '../../api/notifiarrConfig'
import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { profile } from '../../api/profile.svelte'
import type { Command } from '../../api/notifiarrConfig'
import { deepCopy } from '../../includes/util'
import { faTerminal } from '@fortawesome/sharp-duotone-solid-svg-icons'
import { faCommand } from '@fortawesome/sharp-duotone-regular-svg-icons'

export const page = {
  id: 'Commands',
  i: faTerminal,
  c1: 'green',
  c2: 'darkgreen',
  d1: 'seagreen',
  d2: 'green',
}

const empty: Command = {
  name: '',
  hash: '',
  shell: false,
  log: false,
  notify: false,
  args: 0,
}

const merge = (index: number, form: Command): Config => {
  const c = deepCopy(get(profile).config)
  if (!c.commands) c.commands = []
  for (let i = 0; i < c.commands.length; i++) {
    if (i === index) c.commands[i] = form
    else c.commands[i] = {} as Command
  }
  return c
}

const validator = (id: string, value: any): string => {
  id = id.split('.').pop() ?? id
  if (id === 'path') {
    if (!value) return get(_)('FileWatcher.path.required')
  } else if (id === 'regex') {
    if (!value) return get(_)('FileWatcher.regex.required')
  }
  return ''
}

export const app: App<Command> = {
  name: 'Commands',
  id: 'Commands',
  logo: faCommand,
  iconProps: { c1: 'cyan', c2: 'violet', d1: 'violet', d2: 'cyan' },
  disabled: [],
  hidden: [],
  empty,
  merge,
  validator,
}
