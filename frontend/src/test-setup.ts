import '@testing-library/jest-dom/vitest'
import { addMessages, init, waitLocale } from 'svelte-i18n'
import en from './includes/locale/en.json'

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
})

if (!Element.prototype.animate) {
  Element.prototype.animate = () =>
    ({
      cancel() {},
      finish() {},
      pause() {},
      play() {},
      reverse() {},
      addEventListener() {},
      removeEventListener() {},
      dispatchEvent() {
        return true
      },
      finished: Promise.resolve(),
      ready: Promise.resolve(),
    }) as unknown as Animation
}

await addMessages('en', en)
await init({ fallbackLocale: 'en', initialLocale: 'en' })
await waitLocale()
