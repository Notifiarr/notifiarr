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
  private readonly fallbackLocale = 'en'
  // The current locale, so we can export a readonly version.
  private curr = $state(this.fallbackLocale)
  /** Use this to get the current UI locale (translated language). */
  public readonly current = $derived(this.curr)

  constructor() {
    // We only support language codes, not country codes. Maybe one day.
    const init = nav.getQuery('lang') || getLocaleFromNavigator() || this.fallbackLocale
    this.init(init?.split('-')[0], this.fallbackLocale)
  }

  /** Use this to change the UI locale (translated language). */
  public readonly set = async (newLocale: string | null) => {
    if (!newLocale) return

    try {
      newLocale = newLocale.split('-')[0]
      await register(newLocale, async () => await import(`../locale/${newLocale}.json`))
      await lang.set(newLocale)
      await nav.setQuery('lang', (this.curr = newLocale))
    } catch (e) {
      this.error(`Error registering selected locale ${newLocale}: ${e}`)
    }
  }

  private init = async (initial: string, fallback: string) => {
    try {
      await register(initial, async () => await import(`../locale/${initial}.json`))
      await register(fallback, async () => await import(`../locale/${fallback}.json`))
      await init({ fallbackLocale: fallback, initialLocale: initial })
      this.curr = initial
    } catch (e) {
      this.error(`Error registering browser locale ${initial}: ${e}`)
      // Load default locale.
      try {
        await register(fallback, async () => await import(`../locale/${fallback}.json`))
        await init({ fallbackLocale: fallback, initialLocale: (this.curr = fallback) })
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
