import { fetchWithTimeout } from './fetch'
import { profile } from './profile'
import { _, isReady } from '../lib/Translate.svelte'

let loginFailedInvalidCredsMsg = 'Login failed: Invalid username or password'
let loginRequestTimedOutMsg = 'Login request timed out'
let loginFailedMsg = 'Login failed:'

// Translate our messages.
isReady.subscribe(ready => {
  if (ready)
    _.subscribe(val => {
      loginFailedInvalidCredsMsg = val('config.errors.LoginFailedInvalidCreds')
      loginRequestTimedOutMsg = val('config.errors.LoginRequestTimedOut')
      loginFailedMsg = val('config.errors.LoginFailed')
    })
})

export async function login(name: string, password: string): Promise<string | null> {
  try {
    const response = await fetchWithTimeout('?login=true', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({ name, password }),
    })

    if (!response.ok) {
      return loginFailedMsg
    }

    // The login call returns the profile data, so we can set the store and be ready to go.
    // Loading the profile signals to the index page to load the navigation bar.
    profile.set({ ...(await response.json()), loggedIn: true })

    return null
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      return loginRequestTimedOutMsg
    }

    return loginFailedMsg + ' ' + err
  }
}
