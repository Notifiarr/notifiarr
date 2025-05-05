import { writable } from 'svelte/store'
import { fetchWithTimeout } from './util'
import { checkReloaded, postUi } from '../api/fetch'
import type { Config } from '../api/notifiarrConfig'
import { toast } from '@zerodevx/svelte-toast'

interface Profile {
  username: string
  config: Config
}

function createProfileStore() {
  const { subscribe, set, update } = writable<Profile | null>(null)

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
        update(profile => profile ? { ...profile, config } : null)
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

export const profile = createProfileStore()

export async function fetchProfile() {
  try {
    const response = await fetchWithTimeout('/ui/profile')
    if (!response.ok) {
      profile.set(null)
    } else {
      const data = await response.json()
      profile.set(data)
    }
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new Error('Profile request timed out')
    }
    profile.set(null)
    throw error
  }
}

export async function login(name: string, password: string): Promise<string | null> {
  try {
    const response = await fetchWithTimeout('/', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({ name, password }),
    })

    if (!response.ok) {
      return `Login failed: Invalid username or password`
    }

    // The login call returns the profile data, so we can set the store and be ready to go.
    // Loading the profile signals to the index page to load the navigation bar.
    profile.set(await response.json())

    return null
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      return 'Login request timed out'
    }

    return `Login failed: ${err}`
  }
}
