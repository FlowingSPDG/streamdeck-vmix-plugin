import type { inInfo, ActionInfo } from './types/streamdeck'

export type ISD<T> = {
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

  // Callbacks
  callbacks: {
    OnDidReceiveSettings: (f: (settings: T) => void) => void
    OnDidReceiveGlobalSettings: (f: (settings: T) => void) => void
    OnSendToPropertyInspector: (f: (payload: T) => void) => void
  }

  // TODO:
  // logMessage
  // getSettings
  // getGlobalSettings
}

export class SD<T> implements ISD<T> {
  websocket: WebSocket
  uuid: string
  registerEventName: string
  Info: inInfo
  actionInfo: ActionInfo<T>
  runningApps: string[]
  isQT: boolean

  callbacks: {
    OnDidReceiveSettings: (payload: unknown) => void
    OnDidReceiveGlobalSettings: (payload: unknown) => void
    OnSendToPropertyInspector: (payload: unknown) => void
  }

  constructor(
    inPort: number,
    inPropertyInspectorUUID: string,
    inRegisterEvent: string,
    inInfo: string,
    inActionInfo: string,
    // callbacks
    callbacks: {
      OnDidReceiveSettings: (settings: unknown) => void
      OnDidReceiveGlobalSettings: (settings: unknown) => void
      OnSendToPropertyInspector: (settings: unknown) => void
    },
  ) {
    this.websocket = new WebSocket(`ws://127.0.0.1:${inPort}`)
    this.uuid = inPropertyInspectorUUID
    this.registerEventName = inRegisterEvent
    this.Info = JSON.parse(inInfo)
    this.actionInfo = JSON.parse(inActionInfo)
    this.runningApps = []
    this.isQT = navigator.appVersion.includes('QtWebEngine') //  TODO: fix
    this.callbacks = callbacks

    // Register websocket callbacks
    this.websocket.onopen = this.onOpen
    this.websocket.onmessage = this.onMessage

    // Call the plugin to get the current settings
    this.callbacks.OnDidReceiveSettings(this.actionInfo.payload.settings)
  }

  sendValueToPlugin: (value: string, param: string) => void = (value, param) => {
    const json = {
      action: this.actionInfo.action,
      event: 'sendToPlugin',
      context: this.uuid,
      payload: {
        [param]: value,
      },
    }
    this.websocket.send(JSON.stringify(json))
  }

  setSettings: (payload: T) => void = (payload) => {
    const json = {
      event: 'setSettings',
      context: this.uuid,
      payload: payload,
    }
    console.log('Sending payload...', json)
    this.websocket.send(JSON.stringify(json))
  }

  sendPayloadToPlugin: (payload: T) => void = (payload) => {
    const json = {
      action: this.actionInfo.action,
      event: 'sendToPlugin',
      context: this.uuid,
      payload: payload,
    }
    this.websocket.send(JSON.stringify(json))
  }

  openWebsite: (url: string) => void = (url) => {
    const json = {
      event: 'openUrl',
      payload: {
        url: url,
      },
    }
    this.websocket.send(JSON.stringify(json))
  }

  protected onOpen: () => void = () => {
    const json = {
      event: this.registerEventName,
      uuid: this.uuid,
    }
    this.websocket.send(JSON.stringify(json))

    // Notify the plugin that we are connected
    this.sendValueToPlugin('propertyInspectorConnected', 'property_inspector')
  }

  protected onMessage: (event: MessageEvent) => void = (event) => {
    const json = JSON.parse(event.data)
    if (json.event === 'didReceiveSettings') {
      this.callbacks.OnDidReceiveSettings(json.payload)
    }
    if (json.event === 'didReceiveGlobalSettings') {
      this.callbacks.OnDidReceiveGlobalSettings(json.payload)
    }
    if (json.event === 'sendToPropertyInspector') {
      this.callbacks.OnSendToPropertyInspector(json.payload)
    }
  }
}
