import { useState } from 'react'
import { SD } from './sd'
import { Preview, type PreviewSettings } from './components/preview'
import { Program, type ProgramSettings } from './components/program'
import type { DestinationToInputs } from './types/streamdeck'
import { Activator, type ActivatorSettings } from './components/activator'
import { FunctionComponent, type FunctionSettings } from './components/function'
import { TallyMode } from './components/tally'

declare global {
  interface Window {
    connectElgatoStreamDeckSocket: (
      inPort: number,
      inUUID: string,
      inRegisterEvent: string,
      inInfo: string,
      inActionInfo: string
    ) => void
  }
}

function App() {
  type T = PreviewSettings | ProgramSettings | ActivatorSettings | FunctionSettings

  // States
  const [sd, setSD] = useState<SD<unknown> | null>(null)
  const [settings, setSettings] = useState<T | null>(null)
  const [inputs, setInputs] = useState<DestinationToInputs>({})
  const [destinations, setDestinations] = useState<string[]>([])

  // connectElgatoStreamDeckSocket is a function that is called by the Stream Deck software when the Property Inspector is opened.
  // グローバル変数である必要がある
  window.connectElgatoStreamDeckSocket = (
    inPort: number,
    inUUID: string,
    inRegisterEvent: string,
    inInfo: string,
    inActionInfo: string,
  ) => {
    setSD(new SD(inPort, inUUID, inRegisterEvent, inInfo, inActionInfo,
      {
        onOpen: () => {
          console.log('Opened')
        },
        OnDidReceiveSettings: (s: unknown) => {
          console.log('OnDidReceiveSettings', s)
          if (!s || typeof s !== 'object') return

          // sd.tsxから直接設定が渡されるため、settingsプロパティは不要
          setSettings(s as T)
        },
        OnDidReceiveGlobalSettings: (s) => {
          console.log(s)
        },
        OnSendToPropertyInspector: (payload: unknown) => {
          console.log('Received payload', payload)
          if (!payload || typeof payload !== 'object') return
          if (!('event' in payload)) return

          const payloadObj = payload as { event: string }
          if (payloadObj.event === 'inputs') {
            const p = payload as { event: string; inputs: DestinationToInputs }
            console.log('inputs', p.inputs)
            setInputs(p.inputs)
          } else if (payloadObj.event === 'destinations') {
            const p = payload as { event: string; destinations: string[] }
            console.log('destinations', p.destinations)
            setDestinations(p.destinations)
          }
        },
      },
    ))
  }

  const onSettingsUpdate = (s: T) => {
    console.log('onSettingsUpdate', s)
    setSettings(s)
    sd?.setSettings(s)
  }

  // 設定が未設定の場合のデフォルト値を返す
  const getDefaultSettings = (action: string): T => {
    switch (action) {
      case 'dev.flowingspdg.vmix.preview':
        return {
          dest: 'localhost',
          input: 1,
          mix: 0,
          tally_mode: TallyMode.TALLY,
        } as T
      case 'dev.flowingspdg.vmix.program':
        return {
          dest: 'localhost',
          input: 1,
          mix: 0,
          tally_mode: TallyMode.TALLY,
          transition: 'Fade',
          duration: 1000,
        } as T
      case 'dev.flowingspdg.vmix.function':
        return {
          dest: 'localhost',
          function: '',
          query: '',
          input: null,
          acts_event: '',
          acts_input: '',
          acts_active_state: '',
          acts_inactive_state: '',
        } as T
      case 'dev.flowingspdg.vmix.activator':
        return {
          dest: 'localhost',
          input: 1,
          color: 1,
          activator: 'Input',
        } as T
      default:
        return {
          dest: 'localhost',
          input: 1,
          mix: 0,
          tally_mode: TallyMode.TALLY,
        } as T
    }
  }

  if (!sd || !settings) {
    // 初期設定を適用
    if (sd && !settings) {
      setSettings(getDefaultSettings(sd.actionInfo.action))
    }
    return null
  }

  return (
    <>
      { sd.actionInfo.action === 'dev.flowingspdg.vmix.preview' &&
        <Preview {...{
          settings: settings as PreviewSettings,
          inputs,
          destinations,
          onUpdate: onSettingsUpdate,
          sd: sd,
        }} />
      }
      { sd.actionInfo.action === 'dev.flowingspdg.vmix.program' &&
        <Program {...{
          settings: settings as ProgramSettings,
          inputs,
          destinations,
          onUpdate: onSettingsUpdate,
          sd: sd,
        }} />
      }
      { sd.actionInfo.action === 'dev.flowingspdg.vmix.activator' &&
        <Activator {...{
          settings: settings as ActivatorSettings,
          inputs,
          destinations,
          onUpdate: onSettingsUpdate,
          sd: sd,
        }} />
      }
      { sd.actionInfo.action === 'dev.flowingspdg.vmix.function' &&
        <FunctionComponent
          settings={settings as FunctionSettings}
          destinations={destinations}
          onUpdate={onSettingsUpdate}
          sd={sd}
        />
      }
    </>
  )
}

export default App
