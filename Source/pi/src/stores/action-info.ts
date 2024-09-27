import { headlessStreamDeck } from '../adapters/stream-deck'
import { ActionInfo } from '../types/streamdeck'

export const createActionInfoStore = () => {
  let state: ActionInfo<unknown>[] = []
  const listeners = new Set<() => void>()
  const handler = () => {
    state = headlessStreamDeck.getInfos()
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
        headlessStreamDeck.addEventListener('open', handler)
      }

      return () => {
        listeners.delete(callback)

        if (listeners.size === 0) {
          headlessStreamDeck.removeEventListener('open', handler)
        }
      }
    },
  }
}
export const actionInfoStore = createActionInfoStore()
