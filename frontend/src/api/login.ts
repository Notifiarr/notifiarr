import { fetchWithTimeout } from './fetch'
import { profile } from './profile'

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
    profile.set({...(await response.json()), loggedIn: true})

    return null
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      return 'Login request timed out'
    }

    return `Login failed: ${err}`
  }
}
