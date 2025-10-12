import type { App } from '../../includes/formsTracker.svelte'
import type { Config, MySQLConfig } from '../../api/notifiarrConfig'
import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { profile } from '../../api/profile.svelte'
import type { Command } from '../../api/notifiarrConfig'
import { deepCopy } from '../../includes/util'
import { faTerminal, faCommand } from '@fortawesome/sharp-duotone-solid-svg-icons'
import { validate } from '../../includes/instanceValidator'

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
  timeout: '10s',
  command: '',
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

const validator = (
  id: string,
  value: any,
  index: number,
  instances: Command[],
): string => {
  if (id.endsWith('command') && !value) return get(_)('Commands.command.required')
  return validate(id, value, index, instances)
}

export const app: App<Command> = {
  name: 'Commands',
  id: 'Commands',
  envPrefix: 'COMMANDS',
  logo: faCommand,
  iconProps: { c1: 'darksalmon', d1: 'salmon', d2: 'tomato' },
  disabled: [],
  hidden: [],
  empty,
  merge,
  validator,
}
