import { ActionInfo } from '../types/streamdeck'
import { EventListener } from './event-listener'
import { Connection, ConnectParameters, HeadlessStreamDeck, StreamDeckEventMap } from './types'
import ReconnectingWebSocket from 'reconnecting-websocket'

export type StreamDeckOptions = {
  host?: string
  protocol?: 'ws' | 'wss'
  webSocket?: (url: string) => WebSocket
}

const defaultOptions: Required<StreamDeckOptions> = {
  host: '127.0.0.1',
  protocol: 'ws',
  webSocket: url => new ReconnectingWebSocket(url) as WebSocket,
}

export class HeadlessStreamDeckImpl<T> implements HeadlessStreamDeck<T> {
  private readonly connections = new Map<string, Connection<T>>()
  private readonly listeners = new EventListener<StreamDeckEventMap>()
  private readonly options: Required<StreamDeckOptions> = { ...defaultOptions }

  constructor(options: StreamDeckOptions = {}) {
    this.options = { ...this.options, ...options }
  }

  add(inPort: number, options: ConnectParameters): void {
    const urlObj = new URL(`${this.options.protocol}://${this.options.host}:${inPort}`)
    const url = urlObj.toString()

    if (this.connections.has(url)) {
      console.warn(`Port ${inPort} is already in use`)
      return
    }

    const ws = this.options.webSocket(url)

    const actionInfo = JSON.parse(options.inActionInfo)
    this.connections.set(url, {
      uuid: options.inPropertyInspectorUUID,
      registerEventName: options.inRegisterEvent,
      info: JSON.parse(options.inInfo),
      actionInfo,
      isQT: navigator.appVersion.includes('QtWebEngine'),
      ws,
    })
    ws.addEventListener('open', this.onOpen.bind(this))
    ws.addEventListener('message', this.onMessage.bind(this))

    this.listeners.dispatch('didReceiveSettings', actionInfo.payload.settings)
  }

  remove(inPort: number): void {
    const urlObj = new URL(`${this.options.protocol}://${this.options.host}:${inPort}`)
    const url = urlObj.toString()

    const ws = this.connections.get(url)?.ws
    ws?.removeEventListener('open', this.onOpen.bind(this))
    ws?.removeEventListener('message', this.onMessage.bind(this))
    ws?.close()

    this.connections.delete(url)
  }

  getInfo(inPort: number): ActionInfo<T> {
    const urlObj = new URL(`${this.options.protocol}://${this.options.host}:${inPort}`)
    const url = urlObj.toString()

    const conn = this.connections.get(url)
    if (!conn) {
      throw new Error(`Port ${inPort} is not connected`)
    }

    return conn.actionInfo
  }

  getInfos(): ActionInfo<T>[] {
    return Array.from(this.connections.values()).map(conn => conn.actionInfo)
  }

  addEventListener<K extends keyof StreamDeckEventMap>(key: K, callback: StreamDeckEventMap[K]): void {
    this.listeners.add(key, callback)
  }

  removeEventListener<K extends keyof StreamDeckEventMap>(key: K, callback: StreamDeckEventMap[K]): void {
    this.listeners.remove(key, callback)
  }

  sendValueToPlugin(param: string, value: string): void {
    for (const { ws, actionInfo, uuid } of this.connections.values()) {
      const json = {
        action: actionInfo.action,
        event: 'sendToPlugin',
        context: uuid,
        payload: {
          [param]: value,
        },
      }
      ws.send(JSON.stringify(json))
    }
  }

  setSettings(payload: T): void {
    for (const { ws, uuid } of this.connections.values()) {
      const json = {
        event: 'setSettings',
        context: uuid,
        payload,
      }
      ws.send(JSON.stringify(json))
    }

    // internal 向けに先んじてイベントを発火
    this.listeners.dispatch('didReceiveSettings', payload)
  }

  sendPayloadToPlugin(payload: T): void {
    for (const { ws, actionInfo, uuid } of this.connections.values()) {
      const json = {
        action: actionInfo.action,
        event: 'sendToPlugin',
        context: uuid,
        payload,
      }
      ws.send(JSON.stringify(json))
    }
  }

  openWebsite(url: string): void {
    for (const { ws } of this.connections.values()) {
      const json = {
        event: 'openUrl',
        payload: {
          url,
        },
      }
      ws.send(JSON.stringify(json))
    }
  }

  private onOpen() {
    this.listeners.dispatch('open')
  }

  private onMessage(event: MessageEvent) {
    try {
      const parsed = JSON.parse(event.data)
      switch (parsed.event) {
        case 'didReceiveSettings':
          this.listeners.dispatch('didReceiveSettings', parsed.payload.settings)
          break
        case 'didReceiveGlobalSettings':
          this.listeners.dispatch('didReceiveGlobalSettings', parsed.payload.settings)
          break
        case 'sendToPropertyInspector':
          this.listeners.dispatch('sendToPropertyInspector', parsed.payload)
          break
      }
    }
    catch (e) {
      console.error(e)
    }
  }
}
