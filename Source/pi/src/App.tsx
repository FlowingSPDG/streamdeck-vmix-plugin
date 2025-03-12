import { useState } from 'react'
import { SD } from './sd'
import { Preview, type PreviewSettings } from './components/preview'
import { Program, type ProgramSettings } from './components/program'
import type { DestinationToInputs } from './types/streamdeck'
import { Activator, type ActivatorSettings } from './components/activator'
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
  type T = PreviewSettings | ProgramSettings | ActivatorSettings

  // States
  const [sd, setSD] = useState<SD<unknown> | null>(null)
  const getInitialSettings = (action?: string): T => {
    const baseSettings = {
      dest: 'localhost',
      input: 1,
      mix: 0,
      tally_mode: TallyMode.TALLY,
    }

    if (action === 'dev.flowingspdg.vmix.program') {
      return {
        ...baseSettings,
        transition: 'Fade',
        duration: 1000,
      } as T
    }

    return baseSettings as T
  }

  const [settings, setSettings] = useState<T>(getInitialSettings())
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
          
          const settings = (s as T)
          if (!settings || typeof settings !== 'object') return
          
          // 必須フィールドのチェック
          const obj = settings as Record<string, unknown>
          if (
            typeof obj.dest !== 'string' ||
            typeof obj.input !== 'number' ||
            typeof obj.mix !== 'number' ||
            typeof obj.tally_mode !== 'number'
          ) {
            console.error('Invalid settings format:', settings)
            return
          }
          
          setSettings(settings as T)
        },
        OnDidReceiveGlobalSettings: (s) => {
          console.log(s)
        },
        OnSendToPropertyInspector: (payload: unknown) => {
          console.log('Received payload', payload)
          // カスみてえな型チェック
          if (!payload) return
          if (typeof payload !== 'object') return
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

      // TODO: 型をもっと扱いやすく厳密にする
      // Actionごとにカスタムしたくなると思うので、もっと冗長性を持たせる
      // 例えばSettings, コールバック関数を外部から設定できるようにして、StreamDeckとの接続のみを担うコンポーネントを切り出す
      // actionInfo.action で描画先を変更するのではなく、もっと細かく分ける
    ))

    // TODO: Apply colours
    // addDynamicStyles(inInfo.colors);
  }

  const onSettingsUpdate = (s: T) => {
    console.log('onSettingsUpdate', s)
    setSettings(s)
    sd?.setSettings(s)
  }

  return (
    <>
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.preview' && 
        <Preview {...{
          settings: settings as ProgramSettings,
          inputs,
          destinations,
          onUpdate: onSettingsUpdate,
          sd: sd,
        }} />
      }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.program' && 
        <Program {...{
          settings: settings as ProgramSettings,
          inputs,
          destinations,
          onUpdate: onSettingsUpdate,
          sd: sd,
        }} />
      }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.activator' &&  
        <Activator inputs={inputs} settings={settings as ActivatorSettings} onUpdate={onSettingsUpdate} />
      }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.function' && 'NOT YET!' }
    </>
  )
}

export default App
