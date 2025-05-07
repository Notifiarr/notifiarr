import { toast } from '@zerodevx/svelte-toast'
import { onDestroy } from 'svelte'

export function trimPrefix(str: string, prefix: string) {
  return str.slice(str.startsWith(prefix) ? prefix.length : 0)
}

export const success = (m: string) => toast.push(m, {
  theme: {
    '--toastBackground': 'green',
    '--toastColor': 'white',
    '--toastBarBackground': 'olive'
  },

})

export const warning = (m: string) => toast.push(m, {
  theme: {
    '--toastBackground': 'orange',
    '--toastColor': 'white',
    '--toastBarBackground': 'black'
  },
})

export const failure = (m: string) => toast.push(m, {
  theme: {
    '--toastBackground': 'red',
    '--toastColor': 'white',
    '--toastBarBackground': 'royalblue'
  },
})

// onInterval sets an interval and destroys it when it when the page changes.
export function onInterval(callback: () => void, seconds: number) {
  const interval = setInterval(callback, seconds*1000)
  onDestroy(() => clearInterval(interval))

  return interval
}

// onOnce sets a timer and expires it after one invocation.
export function onOnce(callback: () => void, seconds: number) {
  const interval = setInterval(() => {
    clearInterval(interval)
    callback()
  }, seconds*1000)

  return interval
}
