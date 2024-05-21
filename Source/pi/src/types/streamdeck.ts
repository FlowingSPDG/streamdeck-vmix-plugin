export type input = {
  key: string
  name: string
  number: number
}

// StreamDeck
export interface inInfo {
  application: Application
  plugin: Plugin
  devicePixelRatio: number
  colors: Colors
  devices: Device[]
}

export interface Application {
  font: string
  language: string
  platform: string
  platformVersion: string
  version: string
}

export interface Plugin {
  uuid: string
  version: string
}

export interface Colors {
  buttonPressedBackgroundColor: string
  buttonPressedBorderColor: string
  buttonPressedTextColor: string
  disabledColor: string
  highlightColor: string
  mouseDownColor: string
}

export interface Device {
  id: string
  name: string
  size: Size
  type: number
}

export interface Size {
  columns: number
  rows: number
}

export interface ActionInfo<T> {
  action: string
  context: string
  device: string
  payload: Payload<T>
}

export interface Payload<T> {
  settings: T
  coordinates: Coordinates
}

export interface Coordinates {
  column: number
  row: number
}

export interface SendToPropertyInspector<T> {
  event: string
  payload: T
}

export interface SendInputs {
  inputs: DestinationToInputs
}

export interface DestinationToInputs {
  [key:string]: input[]
}
