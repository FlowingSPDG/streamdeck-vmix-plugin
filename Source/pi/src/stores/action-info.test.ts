import { vi } from 'vitest'
import { createActionInfoStore } from './action-info'
import { headlessStreamDeck } from '../adapters/stream-deck'
import { ActionInfo } from '../types/streamdeck'

// headlessStreamDeck をモック
vi.mock('../adapters/stream-deck', () => ({
  headlessStreamDeck: {
    getInfos: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  },
}))

describe('actionInfoStore', () => {
  let mockGetInfos: ReturnType<typeof vi.fn>
  let mockAddEventListener: ReturnType<typeof vi.fn>
  let mockRemoveEventListener: ReturnType<typeof vi.fn>

  beforeEach(() => {
    // モック関数を取得
    mockGetInfos = headlessStreamDeck.getInfos as ReturnType<typeof vi.fn>
    mockAddEventListener = headlessStreamDeck.addEventListener as ReturnType<typeof vi.fn>
    mockRemoveEventListener = headlessStreamDeck.removeEventListener as ReturnType<typeof vi.fn>
  })

  afterEach(() => {
    // モック関数をリセット
    vi.clearAllMocks()
    vi.resetAllMocks()
  })

  it('headlessStreamDeck の "open" イベント時にリスナーが通知を受ける', () => {
    const actionInfoStore = createActionInfoStore()
    const mockListener = vi.fn()

    const mockData: ActionInfo<unknown>[] = [
      {
        action: 'test-action',
        context: 'test-context',
        device: 'test-device',
        payload: { settings: {}, coordinates: { column: 0, row: 0 } },
      },
    ]
    mockGetInfos.mockReturnValue(mockData)

    const unsubscribe = actionInfoStore.subscribe(mockListener)
    const handler = mockAddEventListener.mock.calls[0][1]
    handler()

    expect(mockListener).toHaveBeenCalled()
    expect(actionInfoStore.getValue()).toEqual(mockData)

    unsubscribe()
  })

  it('リスナーの登録時に headlessStreamDeck のイベントリスナーが追加される', () => {
    const actionInfoStore = createActionInfoStore()
    const mockListener = vi.fn()

    // リスナーを登録
    const unsubscribe = actionInfoStore.subscribe(mockListener)

    // headlessStreamDeck.addEventListener が呼ばれたことを確認
    expect(mockAddEventListener).toHaveBeenCalledWith('open', expect.any(Function))

    // リスナーを解除
    unsubscribe()
  })

  it('リスナーの解除時に headlessStreamDeck のイベントリスナーが削除される', () => {
    const actionInfoStore = createActionInfoStore()
    const mockListener = vi.fn()

    const unsubscribe = actionInfoStore.subscribe(mockListener)
    unsubscribe()

    expect(mockRemoveEventListener).toHaveBeenCalledWith('open', expect.any(Function))
  })

  it('リスナー登録時にイベントリスナーが追加され、最後のリスナー解除時に削除される', () => {
    const settingsStore = createActionInfoStore()

    const mockListener1 = vi.fn()
    const mockListener2 = vi.fn()

    const unsubscribe1 = settingsStore.subscribe(mockListener1)
    expect(mockAddEventListener).toHaveBeenCalledTimes(1)

    const unsubscribe2 = settingsStore.subscribe(mockListener2)
    expect(mockAddEventListener).toHaveBeenCalledTimes(1)

    unsubscribe1()
    expect(mockRemoveEventListener).not.toHaveBeenCalled()

    unsubscribe2()
    expect(mockRemoveEventListener).toHaveBeenCalledWith('open', expect.any(Function))
  })

  it('getValue が正しい状態を返す', () => {
    const actionInfoStore = createActionInfoStore()

    const mockData: ActionInfo<unknown>[] = [
      {
        action: 'test-action',
        context: 'test-context',
        device: 'test-device',
        payload: { settings: {}, coordinates: { column: 0, row: 0 } },
      },
    ]
    mockGetInfos.mockReturnValue(mockData)

    const unsubscribe = actionInfoStore.subscribe(() => { })
    const handler = mockAddEventListener.mock.calls[0][1]
    handler()

    expect(actionInfoStore.getValue()).toEqual(mockData)

    unsubscribe()
  })
})
