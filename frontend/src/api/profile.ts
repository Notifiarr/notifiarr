import { get, writable } from 'svelte/store'
import { fetchWithTimeout } from './fetch'
import { checkReloaded, postUi } from './fetch'
import type { Config, Profile } from './notifiarrConfig'
import { toast } from '@zerodevx/svelte-toast'
import { _, isReady } from '../lib/Translate.svelte'
import { failure, success } from '../lib/util'

export const profile = createProfileStore()

let timeoutMsg = 'Profile request timed out'
let unknownErrorMsg = 'An unknown error occurred'
let configUpdateFailedMsg = 'Configuration update failed'
let configurationSaved = 'Configuration saved!'
// Translate our messages.
isReady.subscribe(ready => {
  if (ready)
    _.subscribe(val => {
      timeoutMsg = val('config.errors.ProfileReqTimedOut')
      unknownErrorMsg = val('config.errors.AnUnknownErrorOccurred')
      configUpdateFailedMsg = val('config.errors.ConfigUpdateFailed')
      configurationSaved = val('phrases.ConfigurationSaved')
    })
})

export async function fetchProfile() {
  try {
    const response = await fetchWithTimeout('/ui/profile')
    if (!response.ok) profile.set({} as Profile)
    else {
      const data = await response.json()
      data.loggedIn = true
      profile.set(data)
    }
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError')
      throw new Error(timeoutMsg)
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
        const newConfig = { ...get(profile).config, ...config }
        const response = await postUi('reconfig', JSON.stringify(newConfig), false)
        if (!response) throw new Error(configUpdateFailedMsg)

        // Update the local store with the new config
        update(profile => ({ ...profile, config }))
        success(configurationSaved + ' ' + response)

        // TODO: move both of these into the svelte caller.
        // TODO: so we can visualize the update, reload, check and profile fetch.
        // Wait for the server to reload
        await checkReloaded()
        await fetchProfile()
      } catch (error) {
        failure(configUpdateFailedMsg + ': ' + error)
        throw error instanceof Error ? error : new Error(unknownErrorMsg)
      }
    },
  }
}
