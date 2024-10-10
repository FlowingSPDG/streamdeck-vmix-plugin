import { EventListener } from '../event-listener'

type Events = {
  testEvent: (message: string) => void
  anotherEvent: (value: number) => void
}

describe('EventListener', () => {
  it('should add and dispatch event listeners correctly', () => {
    const listener = new EventListener<Events>()
    const callback = vi.fn()

    listener.add('testEvent', callback)
    listener.dispatch('testEvent', 'Hello')

    expect(callback).toHaveBeenCalledWith('Hello')
  })

  it('should remove event listeners correctly', () => {
    const listener = new EventListener<Events>()
    const callback = vi.fn()

    listener.add('testEvent', callback)
    listener.remove('testEvent', callback)
    listener.dispatch('testEvent', 'Hello')

    expect(callback).not.toHaveBeenCalled()
  })

  it('should not throw when dispatching non-existent event', () => {
    const listener = new EventListener<Events>()

    expect(() => {
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-expect-error
      listener.dispatch('nonExistentEvent', 'Hello')
    }).not.toThrow()
  })

  it('should not throw when removing non-existent callback', () => {
    const listener = new EventListener<Events>()
    const callback = vi.fn()

    expect(() => {
      listener.remove('testEvent', callback)
    }).not.toThrow()
  })

  it('should call each callback once even if added multiple times', () => {
    const listener = new EventListener<Events>()
    const callback = vi.fn()

    listener.add('testEvent', callback)
    listener.add('testEvent', callback)
    listener.dispatch('testEvent', 'Hello')

    expect(callback).toHaveBeenCalledTimes(1)
  })

  it('should call all callbacks for an event', () => {
    const listener = new EventListener<Events>()
    const callback1 = vi.fn()
    const callback2 = vi.fn()

    listener.add('testEvent', callback1)
    listener.add('testEvent', callback2)
    listener.dispatch('testEvent', 'Hello')

    expect(callback1).toHaveBeenCalledWith('Hello')
    expect(callback2).toHaveBeenCalledWith('Hello')
  })
})
