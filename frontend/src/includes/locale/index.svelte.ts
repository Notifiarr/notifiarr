import { init, register, locale as lang, getLocaleFromNavigator } from 'svelte-i18n'
import { failure } from '../util'
import { nav } from '../../navigation/nav.svelte'

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

class Locale {
  // We support English primarily, so make that the default and fallback.
  private fallbackLocale = 'en'
  /** Use this to get the current UI locale (translated language). */
  public current: string = $state(this.fallbackLocale)

  constructor() {
    // We only support language codes, not country codes. Maybe one day.
    const initialLocale = getLocaleFromNavigator()?.split('-')[0] || this.fallbackLocale
    this.init(initialLocale, this.fallbackLocale)
  }

  /** Use this to change the UI locale (translated language). */
  public set = async (newLocale: string | null) => {
    if (!newLocale) return

    try {
      await register(newLocale, async () => await import(`../locale/${newLocale}.json`))
      await lang.set(newLocale)
      await nav.setQuery('lang', (this.current = newLocale))
    } catch (e) {
      this.error(`Error registering selected locale ${newLocale}: ${e}`)
    }
  }

  private init = async (initial: string, fallback: string) => {
    try {
      await register(initial, async () => await import(`../locale/${initial}.json`))
      await init({ fallbackLocale: fallback, initialLocale: initial })
    } catch (e) {
      this.error(`Error registering browser locale ${initial}: ${e}`)
      // Load default locale.
      try {
        await register(fallback, async () => await import(`../locale/${fallback}.json`))
        await init({ fallbackLocale: fallback, initialLocale: fallback })
      } catch (e) {
        this.error(`Error registering default locale ${fallback}: ${e}`)
      }
    }
  }

  private error = (message: string) => {
    console.error(message)
    failure(message)
  }
}

/** Use locale to get the current locale (translation language), or change it. */
export const locale = new Locale()
