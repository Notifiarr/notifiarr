import { profile } from "../../api/profile"
import { init, register, locale } from "svelte-i18n"
import type { Profile } from "../../api/notifiarrConfig"

/*
https://phrase.com/blog/posts/a-step-by-step-guide-to-svelte-localization-with-svelte-i18n-v3/
https://phrase.com/blog/posts/how-to-localize-a-svelte-app-with-svelte-i18n/
https://lokalise.com/blog/svelte-i18n/
*/

// Start with English.
const initialLocale = "en"
register(initialLocale, () => import(`../locale/${initialLocale}.json`))
init({fallbackLocale: initialLocale, initialLocale})

// Keep it up to date in case the user changes the conf.
profile.subscribe((profile: Profile) => {
  if (!profile.config) return // not loaded yet.
  // The ../locale is intentional for vite to work properly.
  register(profile.config.ui.language, () => import(`../locale/${profile.config.ui.language}.json`))
  locale.set(profile.config.ui.language)
})
