import { useState } from 'react'
import { SD } from './sd'
import { Preview, type PreviewSettings } from './components/preview'
import { Program, type ProgramSettings } from './components/program'
import type { SendToPropertyInspector, SendInputs, DestinationToInputs } from './types/streamdeck'
import { Activator, type ActivatorSettings } from './components/activator'

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
  const [settings, setSettings] = useState<T>({} as T)
  const [inputs, setInputs] = useState<DestinationToInputs>({})

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
        OnDidReceiveSettings: (s) => {
          console.log('Settings received', s)
          setSettings(s as T)
        },
        OnDidReceiveGlobalSettings: (s) => {
          console.log(s)
        },
        OnSendToPropertyInspector: (payload: unknown) => {
          // カスみてえな型チェック
          if (!payload) return
          if (typeof payload !== 'object') return
          if (!('event' in payload)) return

          if (payload?.event === 'inputs') {
            const p: SendToPropertyInspector<SendInputs> = payload as SendToPropertyInspector<SendInputs>
            console.log('inputs', p.payload.inputs)
            setInputs(p.payload.inputs)
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
    console.log('Updated. sending payload...', s)
    setSettings(s)
    sd?.setSettings(s)
  }

  return (
    <>
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.preview' && <Preview inputs={inputs} settings={settings as PreviewSettings} onUpdate={onSettingsUpdate} /> }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.program' && <Program inputs={inputs} settings={settings as ProgramSettings} onUpdate={onSettingsUpdate} /> }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.activator' && <Activator inputs={inputs} settings={settings as ActivatorSettings} onUpdate={onSettingsUpdate} /> }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.function' && 'NOT YET!' }
    </>
  )
}

export default App
