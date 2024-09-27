import { headlessStreamDeck } from '../adapters/stream-deck'
import type { Subscriber } from './types'

export const createSettingsStore = () => {
  let state: unknown = {}
  const listeners = new Set<Subscriber<unknown>>()
  const handler = (settings: unknown) => {
    state = settings
    for (const listener of listeners) {
      listener(state)
    }
  }

  return {
    getValue() {
      return state
    },
    subscribe(callback: Subscriber<unknown>) {
      listeners.add(callback)
      if (listeners.size === 1) {
        headlessStreamDeck.addEventListener('didReceiveSettings', handler)
      }

      return () => {
        listeners.delete(callback)

        if (listeners.size === 0) {
          headlessStreamDeck.removeEventListener(
            'didReceiveSettings',
            handler,
          )
        }
      }
    },
  }
}
export const settingsStore = createSettingsStore()
