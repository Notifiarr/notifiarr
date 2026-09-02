import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  afterPageSlide,
  pageSlideFinished,
  pageSlideMs,
  pageSlideStarted,
} from './pageSlide'

describe('pageSlide', () => {
  afterEach(() => {
    vi.useRealTimers()
    pageSlideFinished()
  })

  it('resolves after introend', async () => {
    pageSlideStarted()
    const wait = afterPageSlide()
    pageSlideFinished()
    await expect(wait).resolves.toBeUndefined()
  })

  it('resolves if introend never fires', async () => {
    vi.useFakeTimers()
    pageSlideStarted()
    const wait = afterPageSlide()
    await vi.advanceTimersByTimeAsync(pageSlideMs + 80)
    await expect(wait).resolves.toBeUndefined()
  })

  it('keeps a new slide pending when another starts', async () => {
    vi.useFakeTimers()
    pageSlideStarted()
    const first = afterPageSlide()
    await vi.advanceTimersByTimeAsync(100)
    pageSlideStarted()
    const second = afterPageSlide()
    await vi.advanceTimersByTimeAsync(pageSlideMs + 80 - 100)
    await expect(first).resolves.toBeUndefined()
    let done = false
    void second.then(() => {
      done = true
    })
    await Promise.resolve()
    expect(done).toBe(false)
    pageSlideFinished()
    await expect(second).resolves.toBeUndefined()
  })
})
