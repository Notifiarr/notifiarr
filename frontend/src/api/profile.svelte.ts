import { get, writable, type Unsubscriber, type Writable } from 'svelte/store'
import { checkReloaded, fetchWithTimeout, getUi } from './fetch'
import { postUi } from './fetch'
import type { Config, Profile, ProfilePost } from './notifiarrConfig'
import { _ } from '../includes/Translate.svelte'
import { delay, failure, success } from '../includes/util'
import { urlbase } from './fetch'
import CryptoJS from 'crypto-js'

class ConfigProfile {
  private profile: Writable<Profile>
  public now = $state(Date.now())

  /** Display an error message to the user after calling writeConfig or trustProfile. */
  public error = $state('')
  /** This is the last time the profile configuration was updated. */
  public updated = $state<Date | null>(new Date())

  /** Display a status message (form feedback) to the user after calling one of the update methods. */
  public status = $state('')
  /** Display an error message (form feedback) to the user after calling one of the update methods. */
  public formError = $state('')

  constructor () {
    this.profile = writable<Profile>({} as Profile)
    setInterval(() => (this.now = Date.now()), 1234)
  }

  private set = (value: Profile) => {
    this.updated = new Date()
    // Update local url base in case it changed.
    // The backend will begin using another url base after the reload.
    urlbase.set(value.config?.urlbase ?? '/')
    this.profile.set(value)
  }

  /** Use refresh to refresh the existing profile data from the backend. */
  public async refresh() {
    const { ok, body } = await getUi('profile')
    if (!ok) throw new Error(body)
    body.loggedIn = true
    this.set(body)
  }

  /** Use fetch to set initial profile data when logging in. */
  public async fetch(): Promise<string> {
    const { ok, body } = await getUi('profile')
    if (!ok) {
      this.set({} as Profile)
      return body // error message
    }
    body.loggedIn = true
    this.set(body)
    return ''
  }

  /** Clear the status messages. Use this for the alert toggles and unmount callbacks. */
  public clearStatus() {
    this.status = ''
    this.error = ''
    this.formError = ''
  }

  private async waitForReload() {
    try {
      success(get(_)('phrases.ConfigurationSavedReloading'))
      this.status = get(_)('phrases.Reloading')
      await checkReloaded()
      this.status = get(_)('phrases.UpdatingBackEnd')
      this.error = await this.fetch()
      await (this.error == ''
        ? (this.updated = new Date())
        : (this.error = get(_)('phrases.FailedToReload', {
          values: { error: this.error },
        })))
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
    this.updated = null
    form.password = CryptoJS.MD5(form.password).toString()
    form.newPass = CryptoJS.MD5(form.newPass).toString()

    const { ok, body } = await postUi('profile', JSON.stringify(form), false)

    if (!ok) {
      this.formError = this.error = body
      this.status = ''
    } else {
      await this.waitForReload()
      this.status = ''
    }
  }

  /** Use writeConfig to update a partial configuration on the backend and reload. */
  public async writeConfig(config: Config): Promise<boolean> {
    this.status = get(_)('phrases.SavingConfiguration')
    this.error = ''
    this.updated = null

    // Merge whatever was provided with the existing config.
    const newConfig = { ...get(this.profile).config, ...config }
    // Send the config to the server using postUi.
    const { ok, body } = await postUi('reconfig', JSON.stringify(newConfig), false)
    // If it's an error, set the error message, and exit.
    if (!ok) {
      this.status = ''
      this.formError = this.error = get(_)('config.errors.ConfigUpdateFailed', {
        values: { error: body },
      })

      return false
    }

    // Update the local store with the new config.
    await this.set({ ...get(this.profile), config: newConfig })
    await this.waitForReload()

    return true
  }

  public async setApiKey(apiKey: string): Promise<string | null> {
    try {
      const response = await fetchWithTimeout('?setApiKey=true', {
        method: 'PUT',
        headers: { 'X-Api-Key': apiKey },
      }, 15000)
      if (!response.ok) return `${await response.text()}`
      return null
    } catch (err) {
      return `${err}`
    }
  }

  public async login(name: string, password: string): Promise<string | null> {
    try {
      const sha = CryptoJS.MD5(password).toString()
      const response = await fetchWithTimeout('?login=true', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({ name, password, sha }),
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

  /** Save the tunnel configuration to the backend. */
  public async saveTunnels(primaryTunnel: string, backupTunnel: string) {
    this.status = get(_)('phrases.SavingConfiguration')
    this.error = ''

    const resp = await postUi(
      'tunnel/save',
      JSON.stringify({ PrimaryTunnel: primaryTunnel, BackupTunnel: [backupTunnel] }),
      true,
      10000,
    )

    if (!resp.ok) {
      this.status = ''
      this.formError = resp.body
      return
    }

    this.status = get(_)('phrases.UpdatingBackEnd')
    await delay(1000)
    await this.waitForReload()
    this.status = ''
  }

  public subscribe(run: (value: Profile) => void): Unsubscriber {
    return this.profile.subscribe(run)
  }
}

/** Use profile to get the current profile and configuration data. */
export const profile = new ConfigProfile()
