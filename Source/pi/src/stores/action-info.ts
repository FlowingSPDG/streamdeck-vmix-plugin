import { headlessStreamDeck } from '../adapters/stream-deck'
import type { ActionInfo } from '../types/streamdeck'
import type { Subscriber } from './types'

export const createActionInfoStore = () => {
  let state: ActionInfo<unknown>[] = []
  const listeners = new Set<Subscriber<ActionInfo<unknown>[]>>()
  const handler = () => {
    state = headlessStreamDeck.getInfos()
    for (const listener of listeners) {
      listener(state)
    }
  }

  return {
    getValue() {
      return state
    },
    subscribe(callback: Subscriber<ActionInfo<unknown>[]>) {
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
