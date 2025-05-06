
import { writable } from 'svelte/store'
import { fetchWithTimeout } from './fetch'
import { checkReloaded, postUi } from './fetch'
import type { Config } from './notifiarrConfig'
import { toast } from '@zerodevx/svelte-toast'

interface Profile {
  loggedIn: boolean
  username: string
  config: Config
}

export const profile = createProfileStore()

export async function fetchProfile() {
  try {
    const response = await fetchWithTimeout('/ui/profile')
    if (!response.ok) {
      profile.set({} as Profile)
    } else {
      const data = await response.json()
      data.loggedIn = true
      profile.set(data)
    }
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new Error('Profile request timed out')
    }
    profile.set({} as Profile)
    throw error
  }
}

function createProfileStore() {
  const { subscribe, set, update } = writable<Profile>({} as Profile)

  return {
    subscribe,
    set,
    update,
    async updateConfig(config: Config) {
      try {
        // Send the config to the server using postUi
        const response = await postUi('reconfig', JSON.stringify(config), false)
        if (!response) throw new Error('Failed to save configuration')

        // Update the local store with the new config
        update(profile => ({ ...profile, config }))
        toast.push('Configuration updated: '+response)

        // TODO: move both of these into the svelte caller.
        // TODO: so we can visualize the update, reload, check and profile fetch.
        // Wait for the server to reload
        await checkReloaded()
        await fetchProfile()

      } catch (error) {
        toast.push('Configuration update failed: ' + error)
        throw error instanceof Error ? error : new Error('An unknown error occurred')
      }
    }
  }
}
