import { inInfo, ActionInfo } from '../types/streamdeck'

export interface ISD<T> {
  // Properties
  websocket: WebSocket
  uuid: string
  registerEventName: string
  Info: inInfo
  actionInfo: ActionInfo<T>
  runningApps: string[]
  isQT: boolean

  // Send functions
  sendValueToPlugin: (action: string, context: string, payload: T) => void
  setSettings: (payload: T) => void
  sendPayloadToPlugin: (payload: T) => void // any?
  openWebsite: (url: string) => void

  // TODO:
  // logMessage
  // getSettings
  // getGlobalSettings
}

export type Connection<T> = {
  uuid: string
  registerEventName: string
  info: inInfo
  actionInfo: ActionInfo<T>
  isQT: boolean
  ws: WebSocket
}
export type StreamDeckEventMap = {
  open: () => void
  didReceiveSettings: (settings: unknown) => void
  didReceiveGlobalSettings: (settings: unknown) => void
  sendToPropertyInspector: (settings: unknown) => void
}

export type ConnectParameters = {
  inPropertyInspectorUUID: string
  inRegisterEvent: string
  inInfo: string
  inActionInfo: string
}

export interface HeadlessStreamDeck<T> {
  add(inPort: number, options: ConnectParameters): void
  remove(inPort: number): void
  addEventListener<K extends keyof StreamDeckEventMap, T extends StreamDeckEventMap[K]>(key: K, callback: T): void
  removeEventListener<K extends keyof StreamDeckEventMap, T extends StreamDeckEventMap[K]>(key: K, callback: T): void

  getInfo(inPort: number): ActionInfo<T>
  getInfos(): ActionInfo<T>[]

  // commands
  sendValueToPlugin(action: string, context: string): void
  setSettings(payload: T): void
  sendPayloadToPlugin(payload: T): void
  openWebsite(url: string): void
}
