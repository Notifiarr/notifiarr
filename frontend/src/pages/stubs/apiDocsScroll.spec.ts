import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { lockWindowScroll } from './apiDocsScroll'

describe('lockWindowScroll', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'ResizeObserver',
      class {
        observe() {}
        disconnect() {}
        unobserve() {}
      },
    )
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  const mountHost = () => {
    const el = document.createElement('div')
    document.body.appendChild(el)
    return el
  }

  const mockScroll = (y: number) => {
    Object.defineProperty(window, 'scrollY', {
      value: y,
      configurable: true,
      writable: true,
    })
    return vi.spyOn(window, 'scrollTo').mockImplementation((...args) => {
      const next =
        typeof args[0] === 'number'
          ? args[1]
          : ((args[0] as ScrollToOptions | undefined)?.top ?? 0)
      Object.defineProperty(window, 'scrollY', {
        value: next,
        configurable: true,
        writable: true,
      })
    })
  }

  it('pins the window while an endpoint click is settling', () => {
    vi.useFakeTimers()
    const el = mountHost()
    const scrollTo = mockScroll(240)
    const unlock = lockWindowScroll(el)

    el.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    scrollTo.mockClear()
    Object.defineProperty(window, 'scrollY', {
      value: 0,
      configurable: true,
      writable: true,
    })
    vi.advanceTimersByTime(100)

    expect(scrollTo).toHaveBeenCalled()
    unlock()
    el.remove()
  })

  it('stops pinning after unlock so a later page is not yanked back', () => {
    vi.useFakeTimers()
    const el = mountHost()
    const scrollTo = mockScroll(240)
    const unlock = lockWindowScroll(el)

    el.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    unlock()
    scrollTo.mockClear()
    Object.defineProperty(window, 'scrollY', {
      value: 0,
      configurable: true,
      writable: true,
    })
    vi.advanceTimersByTime(500)

    expect(scrollTo).not.toHaveBeenCalled()
    el.remove()
  })
})
