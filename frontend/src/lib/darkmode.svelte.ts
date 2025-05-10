import { writable } from 'svelte/store'
import Cookies from 'js-cookie'

const cookieName = 'darkmode'

/* We save darkmode in a cookie and a svelte store, so it persists across page loads. */
export const darkMode = writable(false)
export function toggleDarkMode() {
  darkMode.update(currentValue => {
    Cookies.set(cookieName, JSON.stringify(!currentValue))
    return !currentValue
  })
}

const storedValue = Cookies.get('darkmode')
if (storedValue) darkMode.set(JSON.parse(storedValue))
