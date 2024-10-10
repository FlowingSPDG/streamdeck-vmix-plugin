import { vi, Mock } from 'vitest'
import { HeadlessStreamDeckImpl, StreamDeckOptions } from '../headless'
import { MockWebSocket } from './__mock__/websocket'
import { Connection, HeadlessStreamDeck } from '../types'

interface TestHeadLessStreamDeck<T> extends HeadlessStreamDeck<T> {
  connections: Map<string, Connection<T>>
}

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
    vi.stubGlobal('navigator', { appVersion: 'QtWebEngine' })
    vi.spyOn(console, 'warn').mockImplementation(() => { })
    vi.spyOn(console, 'error').mockImplementation(() => { })

    mockWebSocketFactory = vi.fn((url: string) => new MockWebSocket(url))

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
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = headlessStreamDeck.connections
    expect(connections.has(url)).toBe(true)

    expect(didReceiveSettingsCallback).toHaveBeenCalledWith({
      setting1: 'value1',
      setting2: 42,
    })

    expect(mockWebSocketFactory).toHaveBeenCalledWith(url)
  })

  it('同じポートで接続を追加しようとすると警告が表示される', () => {
    headlessStreamDeck.add(options.inPort, options)
    headlessStreamDeck.add(options.inPort, options)

    expect(console.warn).toHaveBeenCalledWith(`Port ${options.inPort} is already in use`)
  })

  it('接続を正しく削除し、WebSocket を閉じる', async () => {
    headlessStreamDeck.add(options.inPort, options)
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = (headlessStreamDeck).connections
    const ws = connections.get(url)?.ws

    headlessStreamDeck.remove(options.inPort)

    expect(connections.has(url)).toBe(false)
    expect(ws?.readyState).toBe(MockWebSocket.CLOSED)
  })

  it('open イベントリスナーが正しく呼び出される', async () => {
    const openCallback = vi.fn()
    headlessStreamDeck.addEventListener('open', openCallback)

    headlessStreamDeck.add(options.inPort, options)
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(openCallback).toHaveBeenCalled()
  })

  it('open 時に初期化 message が送られる', async () => {
    headlessStreamDeck.add(options.inPort, options)
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const conn = headlessStreamDeck.connections.get(url)!
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws = conn.ws as MockWebSocket
    expect(ws.sentMessages).toEqual([
      JSON.stringify({ event: conn.registerEventName, uuid: conn.uuid }),
      JSON.stringify({ action: conn.actionInfo.action, event: 'sendToPlugin', context: conn.uuid, payload: { property_inspector: 'propertyInspectorConnected' } }),
    ])
  })

  it('イベントリスナーを追加・削除し、イベントが適切にディスパッチされる', async () => {
    const didReceiveSettingsCallback = vi.fn()
    headlessStreamDeck.addEventListener('didReceiveSettings', didReceiveSettingsCallback)

    headlessStreamDeck.add(options.inPort, options)
    await new Promise(resolve => setTimeout(resolve, 0))

    headlessStreamDeck.removeEventListener('didReceiveSettings', didReceiveSettingsCallback)

    headlessStreamDeck.add(options.inPort + 1, options)
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(didReceiveSettingsCallback).toHaveBeenCalledTimes(1)
  })

  it('onMessage が正しいイベントをディスパッチする', async () => {
    const didReceiveSettingsCallback = vi.fn()
    headlessStreamDeck.addEventListener('didReceiveSettings', didReceiveSettingsCallback)

    headlessStreamDeck.add(options.inPort, options)
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = headlessStreamDeck.connections
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws = connections.get(url)?.ws as MockWebSocket

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
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(didReceiveSettingsCallback).toHaveBeenCalledWith({
      settings: {
        setting1: 'newValue',
        setting2: 100,
      },
    })
  })

  it('メッセージを正しく送信する', async () => {
    headlessStreamDeck.add(options.inPort, options)
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = headlessStreamDeck.connections
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws = connections.get(url)?.ws as MockWebSocket
    headlessStreamDeck.sendValueToPlugin('param1', 'value1')

    const sentMessage = JSON.parse(ws.sentMessages[2])
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
    await new Promise(resolve => setTimeout(resolve, 0))

    const url = `ws://127.0.0.1:${options.inPort}/`
    const connections = (headlessStreamDeck).connections
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws = connections.get(url)?.ws as MockWebSocket

    const consoleErrorSpy = vi.spyOn(console, 'error')
    ws.dispatchEvent('message', { data: 'invalid JSON' })

    await new Promise(resolve => setTimeout(resolve, 0))

    expect(consoleErrorSpy).toHaveBeenCalled()
  })

  it('DI された WebSocket が使用されることを確認する', () => {
    headlessStreamDeck.add(options.inPort, options)

    const url = `ws://127.0.0.1:${options.inPort}/`

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

    await new Promise(resolve => setTimeout(resolve, 0))

    const newSettings: TestPayload = {
      setting1: 'updatedValue',
      setting2: 99,
    }

    headlessStreamDeck.setSettings(newSettings)

    await new Promise(resolve => setTimeout(resolve, 0))

    expect(didReceiveSettingsCallback).toHaveBeenCalledWith(newSettings)
  })

  it('onOpen は initialized が false の接続のみを初期化する', async () => {
    const options2: ConnectParameters = {
      inPort: 12346,
      inPropertyInspectorUUID: 'test-uuid-2',
      inRegisterEvent: 'registerEvent2',
      inInfo: JSON.stringify({ info: 'test-info-2' }),
      inActionInfo: JSON.stringify({
        action: 'test-action-2',
        context: 'test-context-2',
        device: 'test-device-2',
        payload: {
          settings: {
            setting1: 'value1-2',
            setting2: 84,
          },
        },
      }),
    }

    headlessStreamDeck.add(options.inPort, options)
    headlessStreamDeck.add(options2.inPort, options2)

    const url1 = `ws://127.0.0.1:${options.inPort}/`
    const url2 = `ws://127.0.0.1:${options2.inPort}/`
    const conn1 = headlessStreamDeck.connections.get(url1)!
    const conn2 = headlessStreamDeck.connections.get(url2)!

    headlessStreamDeck.connections.set(url1, { ...conn1, initialized: true })
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws1 = conn1.ws as MockWebSocket
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-expect-error
    const ws2 = conn2.ws as MockWebSocket

    ws1.dispatchEvent('open', { target: ws1 })
    ws2.dispatchEvent('open', { target: ws2 })

    expect(ws1.sentMessages).toEqual([])

    expect(ws2.sentMessages).toEqual([
      JSON.stringify({ event: conn2.registerEventName, uuid: conn2.uuid }),
      JSON.stringify({ action: conn2.actionInfo.action, event: 'sendToPlugin', context: conn2.uuid, payload: { property_inspector: 'propertyInspectorConnected' } }),
    ])

    expect(ws1.sentMessages.length).toBe(0)
    expect(ws2.sentMessages.length).toBe(2)

    expect(headlessStreamDeck.connections.get(url1)?.initialized).toBe(true)
    expect(headlessStreamDeck.connections.get(url2)?.initialized).toBe(true)
  })
})
