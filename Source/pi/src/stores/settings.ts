import { headlessStreamDeck } from '../adapters/stream-deck'

export const createSettingsStore = () => {
  let state: unknown = {}
  const listeners = new Set<() => void>()
  const handler = (settings: unknown) => {
    state = settings
    for (const listener of listeners) {
      listener()
    }
  }

  return {
    getValue() {
      return state
    },
    subscribe(callback: () => void) {
      listeners.add(callback)
      if (listeners.size === 1) {
        headlessStreamDeck.addEventListener('didReceiveSettings', handler)
      }

      return () => {
        listeners.delete(callback)

        if (listeners.size < 1) {
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
