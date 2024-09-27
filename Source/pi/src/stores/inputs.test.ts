import { vi } from 'vitest'
import { createInputsStore } from '.'
import { headlessStreamDeck } from '../adapters/stream-deck'
import { SendToPropertyInspector, SendInputs } from '../types/streamdeck'

// headlessStreamDeck をモック
vi.mock('../adapters/stream-deck', () => ({
  headlessStreamDeck: {
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  },
}))

describe('inputsStore', () => {
  let mockAddEventListener: ReturnType<typeof vi.fn>
  let mockRemoveEventListener: ReturnType<typeof vi.fn>

  beforeEach(() => {
    // モック関数を取得
    mockAddEventListener = headlessStreamDeck.addEventListener as ReturnType<typeof vi.fn>
    mockRemoveEventListener = headlessStreamDeck.removeEventListener as ReturnType<typeof vi.fn>
  })

  afterEach(() => {
    vi.clearAllMocks()
    vi.resetAllMocks()
  })

  it('headlessStreamDeck の "sendToPropertyInspector" イベント時にリスナーが通知を受ける', () => {
    const mockListener = vi.fn()
    const inputsStore = createInputsStore()
    const unsubscribe = inputsStore.subscribe(mockListener)
    const handler = mockAddEventListener.mock.calls[0][1]

    const mockPayload: SendToPropertyInspector<SendInputs> = {
      event: 'inputs',
      payload: {
        inputs: {
          destination1: [{
            key: 'key1',
            name: 'name1',
            number: 1,
          }],
          destination2: [{
            key: 'key2',
            name: 'name2',
            number: 2,
          }],
        },
      },
    }

    handler(mockPayload)
    expect(mockListener).toHaveBeenCalled()
    expect(inputsStore.getValue()).toEqual(mockPayload.payload.inputs)
    unsubscribe()
  })

  it('無効なペイロードではリスナーが通知を受けない', () => {
    const inputsStore = createInputsStore()
    const mockListener = vi.fn()
    const unsubscribe = inputsStore.subscribe(mockListener)
    const handler = mockAddEventListener.mock.calls[0][1]
    const invalidPayloads = [
      null,
      undefined,
      123,
      'string',
      {},
      { event: 'otherEvent' },
    ]

    for (const payload of invalidPayloads) {
      handler(payload)
    }

    expect(mockListener).not.toHaveBeenCalled()
    expect(inputsStore.getValue()).toEqual({})
    unsubscribe()
  })

  it('リスナーの登録時に headlessStreamDeck のイベントリスナーが追加される', () => {
    const mockListener = vi.fn()
    const inputsStore = createInputsStore()

    const unsubscribe = inputsStore.subscribe(mockListener)
    expect(mockAddEventListener).toHaveBeenCalledWith('sendToPropertyInspector', expect.any(Function))
    unsubscribe()
  })

  it('リスナーの解除時に headlessStreamDeck のイベントリスナーが削除される', () => {
    const mockListener = vi.fn()
    const inputsStore = createInputsStore()
    const unsubscribe = inputsStore.subscribe(mockListener)
    unsubscribe()
    expect(mockRemoveEventListener).toHaveBeenCalledWith('sendToPropertyInspector', expect.any(Function))
  })

  it('リスナー登録時にイベントリスナーが追加され、最後のリスナー解除時に削除される', () => {
    const settingsStore = createInputsStore()

    const mockListener1 = vi.fn()
    const mockListener2 = vi.fn()

    const unsubscribe1 = settingsStore.subscribe(mockListener1)
    expect(mockAddEventListener).toHaveBeenCalledTimes(1)

    const unsubscribe2 = settingsStore.subscribe(mockListener2)
    expect(mockAddEventListener).toHaveBeenCalledTimes(1)

    unsubscribe1()
    expect(mockRemoveEventListener).not.toHaveBeenCalled()

    unsubscribe2()
    expect(mockRemoveEventListener).toHaveBeenCalledWith('sendToPropertyInspector', expect.any(Function))
  })

  it('getValue が正しい状態を返す', () => {
    const inputsStore = createInputsStore()
    const unsubscribe = inputsStore.subscribe(() => { })

    const handler = mockAddEventListener.mock.calls[0][1]

    const mockPayload: SendToPropertyInspector<SendInputs> = {
      event: 'inputs',
      payload: {
        inputs: {
          destination1: [],
        },
      },
    }

    handler(mockPayload)
    expect(inputsStore.getValue()).toEqual(mockPayload.payload.inputs)
    unsubscribe()
  })
})
