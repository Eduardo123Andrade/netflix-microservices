import { TimeoutError, withTimeout } from '@/shared/utils'

describe('withTimeout', () => {
  afterEach(() => {
    jest.useRealTimers()
  })

  it('resolves with the promise value when it settles in time', async () => {
    await expect(withTimeout(Promise.resolve('ok'), 100)).resolves.toBe('ok')
  })

  it('propagates the original rejection when it fails in time', async () => {
    const error = new Error('boom')
    await expect(withTimeout(Promise.reject(error), 100)).rejects.toBe(error)
  })

  it('rejects with TimeoutError when the promise does not settle in time', async () => {
    jest.useFakeTimers()
    const pending = withTimeout(new Promise(() => {}), 100)

    jest.advanceTimersByTime(100)

    await expect(pending).rejects.toBeInstanceOf(TimeoutError)
  })

  it('clears the timer once the promise settles', async () => {
    jest.useFakeTimers()

    await withTimeout(Promise.resolve('ok'), 100)

    expect(jest.getTimerCount()).toBe(0)
  })
})
