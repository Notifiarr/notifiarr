import { fetchWithTimeout } from './fetch'
import { profile } from './profile'
import { _, isReady } from '../lib/Translate.svelte'

let invalidCreds = 'Invalid username or password'

// Translate our messages.
isReady.subscribe(ready => {
  if (ready) _.subscribe(val => (invalidCreds = val('config.errors.InvalidCreds')))
})

export async function login(name: string, password: string): Promise<string | null> {
  try {
    const response = await fetchWithTimeout('?login=true', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({ name, password }),
    })

    if (!response.ok) return `${invalidCreds} ${response.status} ${response.statusText}`

    // The login call returns the profile data, so we can set the store and be ready to go.
    // Loading the profile signals to the index page to load the navigation bar.
    profile.set({ ...(await response.json()), loggedIn: true })

    return null
  } catch (err) {
    return `${err}`
  }
}
