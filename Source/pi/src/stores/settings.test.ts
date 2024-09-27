import { vi } from 'vitest'
import { createSettingsStore } from './settings'
import { headlessStreamDeck } from '../adapters/stream-deck'

vi.mock('../adapters/stream-deck', () => ({
  headlessStreamDeck: {
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  },
}))

describe('createSettingsStore', () => {
  let mockAddEventListener: ReturnType<typeof vi.fn>
  let mockRemoveEventListener: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
    mockAddEventListener = headlessStreamDeck.addEventListener as ReturnType<typeof vi.fn>
    mockRemoveEventListener = headlessStreamDeck.removeEventListener as ReturnType<typeof vi.fn>
  })

  it('didReceiveSettings イベント時に状態が更新され、リスナーが通知を受ける', () => {
    const settingsStore = createSettingsStore()

    const mockListener = vi.fn()

    const unsubscribe = settingsStore.subscribe(mockListener)

    expect(mockAddEventListener).toHaveBeenCalledWith('didReceiveSettings', expect.any(Function))

    const handler = mockAddEventListener.mock.calls[0][1]

    const newSettings = { key: 'value' }
    handler(newSettings)

    expect(mockListener).toHaveBeenCalled()
    expect(settingsStore.getValue()).toEqual(newSettings)

    unsubscribe()
  })

  it('リスナー登録時にイベントリスナーが追加され、最後のリスナー解除時に削除される', () => {
    const settingsStore = createSettingsStore()

    const mockListener1 = vi.fn()
    const mockListener2 = vi.fn()

    const unsubscribe1 = settingsStore.subscribe(mockListener1)
    expect(mockAddEventListener).toHaveBeenCalledTimes(1)

    const unsubscribe2 = settingsStore.subscribe(mockListener2)
    expect(mockAddEventListener).toHaveBeenCalledTimes(1)

    unsubscribe1()
    expect(mockRemoveEventListener).not.toHaveBeenCalled()

    unsubscribe2()
    expect(mockRemoveEventListener).toHaveBeenCalledWith('didReceiveSettings', expect.any(Function))
  })

  it('getValue メソッドが正しい状態を返す', () => {
    const settingsStore = createSettingsStore()

    expect(settingsStore.getValue()).toEqual({})

    const mockListener = vi.fn()
    const unsubscribe = settingsStore.subscribe(mockListener)

    const handler = mockAddEventListener.mock.calls[0][1]

    const newSettings = { key: 'value' }
    handler(newSettings)

    expect(settingsStore.getValue()).toEqual(newSettings)

    unsubscribe()
  })
})
