import { get, writable, type Unsubscriber, type Writable } from 'svelte/store'
import { checkReloaded, fetchWithTimeout, getUi } from './fetch'
import { postUi } from './fetch'
import type { Config, Profile, ProfilePost } from './notifiarrConfig'
import { _ } from '../includes/Translate.svelte'
import { failure, success } from '../includes/util'
import { urlbase } from './fetch'

class ConfigProfile {
  private profile: Writable<Profile>
  private now = $state(Date.now())

  /** Display a status message to the user after calling writeConfig or trustProfile. */
  public status = $state('')
  /** Display an error message to the user after calling writeConfig or trustProfile. */
  public error = $state('')
  /** Display a success message to the user after calling writeConfig or trustProfile. */
  public success = $state(<Date | null>null)
  /** Use age to display the age of the success message in milliseconds. */
  public successAge = $derived(this.now - (this.success?.getTime() ?? 0))
  /** This is the last time the profile configuration was updated. */
  public updated = $state(0)
  /** Milliseconds since the last profile configuration update. */
  public updatedAge = $derived(this.now - this.updated)

  constructor() {
    this.profile = writable<Profile>({} as Profile)
    setInterval(() => (this.now = Date.now()), 1234)
  }

  /** Use refresh to refresh the existing profile data. */
  public async refresh() {
    const { ok, body } = await getUi('profile')
    if (!ok) throw new Error(body)
    body.loggedIn = true
    this.set(body)
  }

  /** Use fetch to set initial profile data when logging in. */
  public async fetch() {
    const { ok, body } = await getUi('profile')
    if (!ok) return this.set({} as Profile)
    body.loggedIn = true
    this.set(body)
  }

  /** Clear the status messages. Use this for the alert toggles and unmount callbacks. */
  public clearStatus() {
    this.status = ''
    this.error = ''
    this.success = null
  }

  private async waitForReload() {
    try {
      success(get(_)('phrases.ConfigurationSavedReloading'))
      this.status = get(_)('phrases.Reloading')
      await checkReloaded()
      this.status = get(_)('phrases.UpdatingBackEnd')
      await this.fetch()
      this.success = new Date()
    } catch (e) {
      this.error = `${e}`
      failure(this.error)
    } finally {
      this.status = ''
    }
  }

  /** Use trustProfile to update the authZ/authN configuration on the backend and reload. */
  public async trustProfile(form: ProfilePost) {
    this.status = get(_)('phrases.SavingConfiguration')
    this.error = ''
    this.success = null

    const { ok, body } = await postUi('profile', JSON.stringify(form), false)
    if (!ok) {
      this.error = body
      return (this.status = '')
    }
    await this.waitForReload()
  }

  /** Use writeConfig to update a partial configuration on the backend and reload. */
  public async writeConfig(config: Config) {
    this.status = get(_)('phrases.SavingConfiguration')
    this.error = ''
    this.success = null

    // Merge whatever was provided with the existing config.
    const newConfig = { ...get(this.profile).config, ...config }
    // Send the config to the server using postUi.
    const { ok, body } = await postUi('reconfig', JSON.stringify(newConfig), false)
    // If it's an error, set the error message, and exit.
    if (!ok) {
      this.error = get(_)('config.errors.ConfigUpdateFailed', { values: { error: body } })
      return (this.status = '')
    }

    // Update the local store with the new config.
    await this.set({ ...get(this.profile), config: newConfig })
    await this.waitForReload()
  }

  public async login(name: string, password: string): Promise<string | null> {
    try {
      const response = await fetchWithTimeout('?login=true', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({ name, password }),
      })

      if (!response.ok)
        return `${get(_)('config.errors.InvalidCreds')} ${response.status} ${response.statusText}`

      // The login call returns the profile data, so we can set the store and be ready to go.
      // Loading the profile signals to the index page to load the navigation bar.
      this.set({ ...(await response.json()), loggedIn: true })

      return null
    } catch (err) {
      return `${err}`
    }
  }

  public subscribe(run: (value: Profile) => void): Unsubscriber {
    return this.profile.subscribe(run)
  }

  private set(value: Profile) {
    this.updated = Date.now()
    // Update local url base in case it changed.
    // The backend will begin using another url base after the reload.
    urlbase.set(value.config?.urlbase ?? '/')
    this.profile.set(value)
  }
}

/** Use profile to get the current profile and configuration data. */
export const profile = new ConfigProfile()
