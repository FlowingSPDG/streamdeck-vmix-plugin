import { vi, Mock } from 'vitest'
import { HeadlessStreamDeckImpl, StreamDeckOptions } from '../headless'
import { MockWebSocket } from './__mock__/websocket'
import { Connection, HeadlessStreamDeck } from '../types'

interface TestHeadLessStreamDeck<T> extends HeadlessStreamDeck<T> {
  connections: Map<string, Connection<T>>
}

// 必要な型やインターフェースの定義
interface TestPayload {
  setting1: string
  setting2: number
}

interface ConnectParameters {
  inPort: number
  inPropertyInspectorUUID: string
  inRegisterEvent: string
  inInfo: string
  inActionInfo: string
}

// モック WebSocket クラス

describe('HeadlessStreamDeckImpl', () => {
  let headlessStreamDeck: TestHeadLessStreamDeck<TestPayload>
  const options: ConnectParameters = {
    inPort: 12345,
    inPropertyInspectorUUID: 'test-uuid',
    inRegisterEvent: 'registerEvent',
    inInfo: JSON.stringify({ info: 'test-info' }),
    inActionInfo: JSON.stringify({
      action: 'test-action',
      context: 'test-context',
      device: 'test-device',
      payload: {
        settings: {
          setting1: 'value1',
          setting2: 42,
        },
      },
    }),
  }
  let mockWebSocketFactory: Mock

  beforeEach(() => {
    // グローバルオブジェクトをスタブ
    vi.stubGlobal('navigator', { appVersion: 'QtWebEngine' })

    // コンソールメソッドをモック
    vi.spyOn(console, 'warn').mockImplementation(() => { })
    vi.spyOn(console, 'error').mockImplementation(() => { })

    // MockWebSocket を返すファクトリ関数を作成
    mockWebSocketFactory = vi.fn((url: string) => new MockWebSocket(url))

    // HeadlessStreamDeckImpl のインスタンスを作成し、webSocket オプションにモックを注入
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    headlessStreamDeck = new HeadlessStreamDeckImpl<TestPayload>({
      webSocket: mockWebSocketFactory,
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
    vi.resetAllMocks()
  })

  it('接続を正しく追加し、didReceiveSettings イベントをディスパッチする', async () => {
    const didReceiveSettingsCallback = vi.fn()
    headlessStreamDeck.addEventListener('didReceiveSettings', didReceiveSettingsCallback)

    headlessStreamDeck.add(options.inPort, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    // 接続が追加されたことを確認
    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = headlessStreamDeck.connections
    expect(connections.has(url)).toBe(true)

    // didReceiveSettings イベントが正しくディスパッチされたことを確認
    expect(didReceiveSettingsCallback).toHaveBeenCalledWith({
      setting1: 'value1',
      setting2: 42,
    })

    // WebSocket ファクトリが正しく呼び出されたことを確認
    expect(mockWebSocketFactory).toHaveBeenCalledWith(url)
  })

  it('同じポートで接続を追加しようとすると警告が表示される', () => {
    headlessStreamDeck.add(options.inPort, options)

    // 同じポートで再度追加
    headlessStreamDeck.add(options.inPort, options)

    // 警告が表示されたことを確認
    expect(console.warn).toHaveBeenCalledWith(`Port ${options.inPort} is already in use`)
  })

  it('接続を正しく削除し、WebSocket を閉じる', async () => {
    headlessStreamDeck.add(options.inPort, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = (headlessStreamDeck).connections
    const ws = connections.get(url)?.ws

    // 接続を削除
    headlessStreamDeck.remove(options.inPort)

    expect(connections.has(url)).toBe(false)
    expect(ws?.readyState).toBe(MockWebSocket.CLOSED)
  })

  it('open イベントリスナーが正しく呼び出される', async () => {
    const openCallback = vi.fn()
    Object.defineProperty(openCallback, 'name', { value: 'hoge' })
    headlessStreamDeck.addEventListener('open', openCallback)

    headlessStreamDeck.add(options.inPort, options)

    // WebSocket の open イベントを待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    // open イベントが正しくディスパッチされたことを確認
    expect(openCallback).toHaveBeenCalled()
  })

  it('イベントリスナーを追加・削除し、イベントが適切にディスパッチされる', async () => {
    const didReceiveSettingsCallback = vi.fn()
    headlessStreamDeck.addEventListener('didReceiveSettings', didReceiveSettingsCallback)

    headlessStreamDeck.add(options.inPort, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    // リスナーを削除
    headlessStreamDeck.removeEventListener('didReceiveSettings', didReceiveSettingsCallback)

    // 再度 add を呼び出してもコールバックが呼ばれないことを確認
    headlessStreamDeck.add(options.inPort + 1, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(didReceiveSettingsCallback).toHaveBeenCalledTimes(1)
  })

  it('onMessage が正しいイベントをディスパッチする', async () => {
    const didReceiveSettingsCallback = vi.fn()
    headlessStreamDeck.addEventListener('didReceiveSettings', didReceiveSettingsCallback)

    headlessStreamDeck.add(options.inPort, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = headlessStreamDeck.connections
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws = connections.get(url)?.ws as MockWebSocket

    // メッセージをシミュレート
    const messageData = JSON.stringify({
      event: 'didReceiveSettings',
      payload: {
        settings: {
          setting1: 'newValue',
          setting2: 100,
        },
      },
    })

    ws.dispatchEvent('message', { data: messageData })

    // メッセージの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(didReceiveSettingsCallback).toHaveBeenCalledWith({
      setting1: 'newValue',
      setting2: 100,
    })
  })

  it('メッセージを正しく送信する', async () => {
    headlessStreamDeck.add(options.inPort, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = (headlessStreamDeck).connections
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws = connections.get(url)?.ws as MockWebSocket

    // sendValueToPlugin を呼び出し
    headlessStreamDeck.sendValueToPlugin('param1', 'value1')

    expect(ws.sentMessages.length).toBe(1)
    const sentMessage = JSON.parse(ws.sentMessages[0])
    expect(sentMessage).toEqual({
      action: 'test-action',
      event: 'sendToPlugin',
      context: 'test-uuid',
      payload: {
        param1: 'value1',
      },
    })
  })

  it('正しい ActionInfo を取得する', async () => {
    headlessStreamDeck.add(options.inPort, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    const actionInfo = headlessStreamDeck.getInfo(options.inPort)
    expect(actionInfo).toEqual(JSON.parse(options.inActionInfo))
  })

  it('存在しないポートに対して getInfo を呼び出すとエラーが発生する', () => {
    expect(() => {
      headlessStreamDeck.getInfo(9999)
    }).toThrow('Port 9999 is not connected')
  })

  it('無効な JSON を受信したときにエラーを適切に処理する', async () => {
    headlessStreamDeck.add(options.inPort, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = (headlessStreamDeck).connections
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws = connections.get(url)?.ws as MockWebSocket

    const consoleErrorSpy = vi.spyOn(console, 'error')

    // 無効な JSON をシミュレート
    ws.dispatchEvent('message', { data: 'invalid JSON' })

    // メッセージの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(consoleErrorSpy).toHaveBeenCalled()
  })

  it('DI された WebSocket が使用されることを確認する', () => {
    headlessStreamDeck.add(options.inPort, options)

    const url = `ws://127.0.0.1:${options.inPort}/`

    // WebSocket ファクトリが正しく呼び出されたことを確認
    expect(mockWebSocketFactory).toHaveBeenCalledWith(url)
  })

  it('StreamDeckOptions の host と protocol が正しく適用される', () => {
    const customOptions = {
      host: 'localhost',
      protocol: 'wss',
      webSocket: mockWebSocketFactory,
    } satisfies StreamDeckOptions

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    headlessStreamDeck = new HeadlessStreamDeckImpl<TestPayload>(customOptions)
    headlessStreamDeck.add(options.inPort, options)

    const url = `wss://localhost:${options.inPort}/`

    expect(mockWebSocketFactory).toHaveBeenCalledWith(url)
  })

  it('setSettings を呼び出したときに didReceiveSettings イベントがディスパッチされる', async () => {
    const didReceiveSettingsCallback = vi.fn()
    headlessStreamDeck.addEventListener('didReceiveSettings', didReceiveSettingsCallback)

    headlessStreamDeck.add(options.inPort, options)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    // 新しい設定値
    const newSettings: TestPayload = {
      setting1: 'updatedValue',
      setting2: 99,
    }

    // setSettings を呼び出す
    headlessStreamDeck.setSettings(newSettings)

    // イベントの処理を待つ
    await new Promise(resolve => setTimeout(resolve, 0))

    // didReceiveSettings イベントが正しくディスパッチされたことを確認
    expect(didReceiveSettingsCallback).toHaveBeenCalledWith(newSettings)
  })
})
