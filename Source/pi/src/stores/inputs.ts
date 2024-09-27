import { headlessStreamDeck } from '../adapters/stream-deck'
import { DestinationToInputs, SendToPropertyInspector, SendInputs } from '../types/streamdeck'

export const createInputsStore = () => {
  let state: DestinationToInputs = {}
  const listeners = new Set<() => void>()
  const handler = (payload: unknown) => {
    if (!payload) return
    if (typeof payload !== 'object') return
    if (!('event' in payload)) return
    if (payload?.event !== 'inputs') return
    console.log(payload,

    )

    const p: SendToPropertyInspector<SendInputs> = payload as SendToPropertyInspector<SendInputs>
    state = p.payload.inputs
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
        headlessStreamDeck.addEventListener(
          'sendToPropertyInspector',
          handler,
        )
      }

      return () => {
        listeners.delete(callback)

        if (listeners.size === 0) {
          headlessStreamDeck.removeEventListener(
            'sendToPropertyInspector',
            handler,
          )
        }
      }
    },
  }
}
export const inputsStore = createInputsStore()
