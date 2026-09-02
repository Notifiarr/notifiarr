/** Pin the window while RapiDoc expands an endpoint so the page does not jump. */
export const lockWindowScroll = (el: HTMLElement): (() => void) => {
  let cancelled = false
  let raf = 0
  let pinned = window.scrollY
  let freeze = 0
  const html = document.documentElement
  const prevAnchor = html.style.overflowAnchor
  html.style.overflowAnchor = 'none'

  const inside = (node: Node | null) => {
    if (!node) return false
    const root = node.getRootNode()
    return root === el.shadowRoot || el.contains(node) || node === el
  }

  const restore = () => {
    if (cancelled) return
    if (window.scrollY !== pinned) window.scrollTo(0, pinned)
  }

  const onUserScroll = () => {
    if (freeze) {
      restore()
      return
    }
    pinned = window.scrollY
  }
  window.addEventListener('scroll', onUserScroll, { passive: true })

  const origIntoView = Element.prototype.scrollIntoView
  Element.prototype.scrollIntoView = function (
    this: Element,
    arg?: boolean | ScrollIntoViewOptions,
  ) {
    if (inside(this)) return
    origIntoView.call(this, arg as boolean)
  }

  const origFocus = HTMLElement.prototype.focus
  HTMLElement.prototype.focus = function (this: HTMLElement, opts?: FocusOptions) {
    origFocus.call(this, inside(this) ? { ...opts, preventScroll: true } : opts)
  }

  const hold = () => {
    if (cancelled) return
    freeze++
    pinned = freeze === 1 ? window.scrollY : pinned
    el.style.minHeight = `${el.offsetHeight}px`
    restore()
    const start = performance.now()
    const tick = () => {
      if (cancelled) return
      restore()
      if (performance.now() - start < 400) {
        raf = requestAnimationFrame(tick)
        return
      }
      el.style.minHeight = ''
      freeze--
    }
    raf = requestAnimationFrame(tick)
  }

  el.addEventListener('click', hold, true)
  el.addEventListener('toggle', hold, true)

  const ro = new ResizeObserver(restore)
  ro.observe(el)

  return () => {
    cancelled = true
    cancelAnimationFrame(raf)
    window.removeEventListener('scroll', onUserScroll)
    Element.prototype.scrollIntoView = origIntoView
    HTMLElement.prototype.focus = origFocus
    el.removeEventListener('click', hold, true)
    el.removeEventListener('toggle', hold, true)
    ro.disconnect()
    el.style.minHeight = ''
    html.style.overflowAnchor = prevAnchor
  }
}
