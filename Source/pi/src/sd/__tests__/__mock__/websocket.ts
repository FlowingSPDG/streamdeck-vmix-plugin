import { EventListener } from '../../event-listener'

type EventMap = {
  open: (event: { target: MockWebSocket }) => void
  message: (event: { data: string }) => void
  close: (event: { target: MockWebSocket }) => void
}

export class MockWebSocket {
  readyState: number = 0 // CONNECTING
  sentMessages: string[] = []
  // EventListener を使用
  private listeners = new EventListener<EventMap>()

  static CONNECTING = 0 as const
  static OPEN = 1 as const
  static CLOSING = 2 as const
  static CLOSED = 3 as const

  constructor(readonly url: string) {
    // 接続をシミュレート
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN
      this.dispatchEvent('open', { target: this })
    }, 0)
  }

  addEventListener<K extends keyof EventMap>(event: K, callback: EventMap[K]) {
    this.listeners.add(event, callback)
  }

  removeEventListener<K extends keyof EventMap>(event: K, callback: EventMap[K]) {
    this.listeners.remove(event, callback)
  }

  dispatchEvent<K extends keyof EventMap>(event: K, ...data: Parameters<EventMap[K]>) {
    this.listeners.dispatch(event, ...data)
  }

  send(data: string) {
    this.sentMessages.push(data)
  }

  close() {
    this.readyState = MockWebSocket.CLOSED
    this.dispatchEvent('close', { target: this })
  }
}
