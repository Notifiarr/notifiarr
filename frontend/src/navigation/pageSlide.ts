/** Matches Svelte's default `transition:slide` duration. */
export const pageSlideMs = 400

const fallbackMs = pageSlideMs + 80

let finish: (() => void) | undefined
let current: Promise<void> = Promise.resolve()

/** Call before swapping `nav.ActivePage` so lazy pages can wait out the intro. */
export const pageSlideStarted = () => {
  let settled = false
  let done!: () => void
  const resolve = () => {
    if (settled) return
    settled = true
    if (finish === resolve) finish = undefined
    done()
  }
  current = new Promise<void>(resolvePromise => {
    done = resolvePromise
  })
  finish = resolve
  setTimeout(resolve, fallbackMs)
}

/** `{#key}` intro finished — safe to mount heavy page content. */
export const pageSlideFinished = () => finish?.()

/** Resolves when the current page intro ends (or the fallback timer). */
export const afterPageSlide = () => current
