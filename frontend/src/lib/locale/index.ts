import { profile } from '../../api/profile'
import { init, register, locale, getLocaleFromNavigator } from 'svelte-i18n'
import type { Profile } from '../../api/notifiarrConfig'
import { failure } from '../util'

/*
https://phrase.com/blog/posts/a-step-by-step-guide-to-svelte-localization-with-svelte-i18n-v3/
https://phrase.com/blog/posts/how-to-localize-a-svelte-app-with-svelte-i18n/
https://lokalise.com/blog/svelte-i18n/
*/

export async function setLocale(newLocale: string) {
  try {
    await register(newLocale, async () => await import(`../locale/${newLocale}.json`))
    await locale.set(newLocale)
  } catch (e) {
    console.error(`Error registering selected locale ${newLocale}:`, e)
    failure(`Error registering selected locale ${newLocale}: ${e}`)
  }
}

// We support English primarily, so make that the default and fallback.
const fallbackLocale = 'en'
// We only support language codes, not country codes. Maybe one day.
const initialLocale = getLocaleFromNavigator()?.split('-')[0] || fallbackLocale

async function initLocale() {
  try {
    await register(
      initialLocale,
      async () => await import(`../locale/${initialLocale}.json`),
    )
    await init({ fallbackLocale, initialLocale })
  } catch (e) {
    failure(`Error registering browser locale ${initialLocale}: ${e}`)
    console.error(`Error registering browser locale ${initialLocale}:`, e)
    // Load default locale.
    register(fallbackLocale, () => import(`../locale/${fallbackLocale}.json`))
    init({ fallbackLocale, initialLocale: fallbackLocale })
  }
}

initLocale()
