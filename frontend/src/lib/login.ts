import { writable } from 'svelte/store'
import { fetchWithTimeout } from './util'

interface Profile {
  username: string
}

export const profile = writable<Profile | null>(null)

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

    // Loading the profile signals to the index page to load the navigation bar.
    await fetchProfile()

    return null
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      return 'Login request timed out'
    }

    return `Login failed: ${(err)}`
  }
}
