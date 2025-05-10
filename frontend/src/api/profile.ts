import { get, writable, type Unsubscriber } from 'svelte/store'
import { getUi } from './fetch'
import { checkReloaded, postUi } from './fetch'
import type { Config, Profile } from './notifiarrConfig'
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

export async function updateProfile() {
  const { ok, body } = await getUi('profile')
  if (!ok) throw new Error(body)
  body.loggedIn = true
  profile.set(body)
}

export async function fetchProfile() {
  const { ok, body } = await getUi('profile')
  if (!ok) return profile.set({} as Profile)
  body.loggedIn = true
  profile.set(body)
}

function createProfileStore() {
  const { subscribe, set, update } = writable<Profile>({} as Profile)
  return { subscribe, set, update, writeConfig }
}

async function writeConfig(config: Config) {
  try {
    // Send the config to the server using postUi
    const newConfig = { ...get(profile).config, ...config }
    const { ok, body } = await postUi('reconfig', JSON.stringify(newConfig), false)
    if (!ok) throw new Error(`${configUpdateFailedMsg}: ${body}`)

    // Update the local store with the new config
    profile.update(profile => ({ ...profile, config }))
    success(configurationSaved + ' ' + body)
  } catch (error) {
    failure(configUpdateFailedMsg + ': ' + error)
    throw error instanceof Error ? error : new Error(unknownErrorMsg)
  }
}
