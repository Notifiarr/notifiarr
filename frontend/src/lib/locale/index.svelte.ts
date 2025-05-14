import { init, register, locale, getLocaleFromNavigator } from 'svelte-i18n'
import { failure } from '../util'

/*
https://phrase.com/blog/posts/a-step-by-step-guide-to-svelte-localization-with-svelte-i18n-v3/
https://phrase.com/blog/posts/how-to-localize-a-svelte-app-with-svelte-i18n/
https://lokalise.com/blog/svelte-i18n/
*/

/**
 * Flags are the flags for the languages.
 * They are close to a country, but not exactly.
 * For example, "es" is "🇪🇸" for Spain, but "🇲🇽" for Mexico.
 */
export const Flags: Record<string, string> = {
  de: '🇩🇪',
  en: '🇺🇸',
  es: '🇲🇽',
  fi: '🇫🇮',
  fr: '🇫🇷',
  hu: '🇭🇺',
  it: '🇮🇹',
  nl: '🇳🇱',
  pl: '🇵🇱',
  pt: '🇵🇹',
  sv: '🇸🇪',
  zh_Hant: '🇹🇼',
  zh_Hans: '🇨🇳',
}

// We support English primarily, so make that the default and fallback.
const fallbackLocale = 'en'
// We only support language codes, not country codes. Maybe one day.
const initialLocale = getLocaleFromNavigator()?.split('-')[0] || fallbackLocale

let current = $state(initialLocale)

export const currentLocale = () => current

export async function setLocale(newLocale: string) {
  try {
    await register(newLocale, async () => await import(`../locale/${newLocale}.json`))
    await locale.set(newLocale)
    current = newLocale
    // Update the URL with the new locale.
    const query = new URLSearchParams(window.location.search)
    await query.set('lang', newLocale)
    window.history.replaceState({}, '', `${window.location.pathname}?${query.toString()}`)
  } catch (e) {
    console.error(`Error registering selected locale ${newLocale}:`, e)
    failure(`Error registering selected locale ${newLocale}: ${e}`)
  }
}

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
