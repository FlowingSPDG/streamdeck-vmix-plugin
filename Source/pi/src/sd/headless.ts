import { ActionInfo } from '../types/streamdeck'
import { Connection, ConnectParameters, HeadlessStreamDeck, StreamDeckEventMap } from './types'
import ReconnectingWebSocket from 'reconnecting-websocket'

export class HeadlessStreamDeckImpl<T> implements HeadlessStreamDeck<T> {
  private readonly connections = new Map<string, Connection<T>>()
  private readonly openCallbacks = new Set<() => void>()
  private readonly didReceiveSettingsCallbacks = new Set<(settings: unknown) => void>()
  private readonly didReceiveGlobalSettingsCallbacks = new Set<(settings: unknown) => void>()
  private readonly sendToPropertyInspectorCallbacks = new Set<(settings: unknown) => void>()

  add(inPort: number, options: ConnectParameters): void {
    const url = `ws://127.0.0.1:${inPort}`
    if (this.connections.has(url)) {
      console.warn(`Port ${inPort} is already in use`)
      return
    }

    // ReconnectingWebSocket はライブラリ側の型定義がミスってるが WebSocket 互換なのでキャストする
    const ws = new ReconnectingWebSocket(url) as WebSocket

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

    for (const callback of this.didReceiveSettingsCallbacks) {
      callback(actionInfo.payload.settings)
    }
  }

  remove(inPort: number): void {
    const url = `ws://127.0.0.1:${inPort}`

    const ws = this.connections.get(url)?.ws
    ws?.removeEventListener('open', this.onOpen.bind(this))
    ws?.removeEventListener('message', this.onMessage.bind(this))
    ws?.close()

    this.connections.delete(url)
  }

  getInfo(inPort: number): ActionInfo<T> {
    const url = `ws://127.0.0.1:${inPort}`
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
    switch (key) {
      case 'open':
        this.openCallbacks.add(callback as () => void)
        break
      case 'didReceiveSettings':
        this.didReceiveSettingsCallbacks.add(callback as (settings: unknown) => void)
        break
      case 'didReceiveGlobalSettings':
        this.didReceiveGlobalSettingsCallbacks.add(callback as (settings: unknown) => void)
        break
      case 'sendToPropertyInspector':
        this.sendToPropertyInspectorCallbacks.add(callback as (settings: unknown) => void)
        break
    }
  }

  removeEventListener<K extends keyof StreamDeckEventMap>(key: K, callback: StreamDeckEventMap[K]): void {
    switch (key) {
      case 'open':
        this.openCallbacks.delete(callback as () => void)
        break
      case 'didReceiveSettings':
        this.didReceiveSettingsCallbacks.delete(callback as (settings: unknown) => void)
        break
      case 'didReceiveGlobalSettings':
        this.didReceiveGlobalSettingsCallbacks.delete(callback as (settings: unknown) => void)
        break
      case 'sendToPropertyInspector':
        this.sendToPropertyInspectorCallbacks.delete(callback as (settings: unknown) => void)
        break
    }
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
    for (const callback of this.openCallbacks) {
      callback()
    }
  }

  private onMessage(event: MessageEvent) {
    try {
      const parsed = JSON.parse(event.data)
      switch (parsed.event) {
        case 'didReceiveSettings': {
          for (const callback of this.didReceiveSettingsCallbacks) {
            callback(parsed.payload.settings)
          }
          break
        }
        case 'didReceiveGlobalSettings': {
          for (const callback of this.didReceiveGlobalSettingsCallbacks) {
            callback(parsed.payload.settings)
          }
          break
        }
        case 'sendToPropertyInspector': {
          for (const callback of this.sendToPropertyInspectorCallbacks) {
            callback(parsed.payload)
          }
          break
        }
      }
    }
    catch (e) {
      console.error(e)
    }
  }
}
