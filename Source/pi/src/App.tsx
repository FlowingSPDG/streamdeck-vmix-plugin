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
  const [settings, setSettings] = useState<T | undefined>(undefined)
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
          setSettings(s as T)
        },
        OnDidReceiveGlobalSettings: (s) => {
          console.log(s)
        },
        OnSendToPropertyInspector: (payload: unknown) => {
          console.log('Received payload', payload)
          // カスみてえな型チェック
          if (!payload) return
          if (typeof payload !== 'object') return
          if (!('payload' in payload)) return

          const payloadObj = payload as { payload: { event: string } }
          if (payloadObj.payload.event === 'inputs') {
            const p = payload as { payload: { event: string; inputs: DestinationToInputs } }
            console.log('inputs', p.payload.inputs)
            setInputs(p.payload.inputs)
          } else if (payloadObj.payload.event === 'destinations') {
            const p = payload as { payload: { event: string; destinations: string[] } }
            console.log('destinations', p.payload.destinations)
            setDestinations(p.payload.destinations)
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
          settings: {
            dest: (settings as PreviewSettings).dest ?? 'localhost',
            input: (settings as PreviewSettings).input ?? 1,
            mix: (settings as PreviewSettings).mix ?? 0,
            tally_mode: (settings as PreviewSettings).tally_mode ?? TallyMode.TALLY,
          },
          inputs,
          destinations,
          onUpdate: onSettingsUpdate,
          sd: sd,
        }} />
      }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.program' && 
        <Program {...{
          settings: {
            dest: (settings as ProgramSettings).dest ?? 'localhost',
            input: (settings as ProgramSettings).input ?? 1,
            mix: (settings as ProgramSettings).mix ?? 0,
            tally_mode: (settings as ProgramSettings).tally_mode ?? TallyMode.TALLY,
            transition: (settings as ProgramSettings).transition ?? 'Fade',
            duration: (settings as ProgramSettings).duration ?? 1000,
          },
          inputs,
          destinations,
          onUpdate: onSettingsUpdate,
          sd: sd,
        }} />
      }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.activator' && <Activator inputs={inputs} settings={settings as ActivatorSettings} onUpdate={onSettingsUpdate} /> }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.function' && 'NOT YET!' }
    </>
  )
}

export default App
