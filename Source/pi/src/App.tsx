import { useState } from 'react'
import { SD } from './sd'
import { Preview, type PreviewSettings } from './components/preview'
import type { input } from './types/vmix'
import type { SendToPropertyInspector, SendInputs } from './types/streamdeck'

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
  // States
  const [sd, setSD] = useState<SD<unknown> | null>(null)
  const [settings, setSettings] = useState<PreviewSettings>({
    host: 'localhost',
    port: 8088,
    input: '',
    tally: true,
    mix: 0,
  })
  const [inputs, setInputs] = useState<input[]>([])

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
        OnDidReceiveSettings: (settings) => {
          setSettings(settings as PreviewSettings)
        },
        OnDidReceiveGlobalSettings: (settings) => {
          console.log(settings)
        },
        OnSendToPropertyInspector: (payload: unknown) => {
          // カスみてえな型チェック
          if (!payload) return
          if (typeof payload !== 'object') return
          if (!('event' in payload)) return

          if (payload?.event === 'inputs') {
            const p: SendToPropertyInspector<SendInputs> = payload as SendToPropertyInspector<SendInputs>
            setInputs(p.payload.inputs)
          }
        },
      },
    ))

    // TODO: Apply colours
    // addDynamicStyles(inInfo.colors);
  }

  // ファイルを変えるのではなく、入ってくるactionに応じてここで何を描画するか切り替えてもいいかもしれない?
  const onUpdate = (settings: PreviewSettings) => {
    console.log('Updated. sending payload...', settings)
    setSettings(settings)
    sd?.setSettings(settings)
  }

  return (
    <>
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.preview' && <Preview inputs={inputs} settings={settings} onUpdate={onUpdate} /> }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.program' && 'NOT YET!' }
      { sd?.actionInfo.action === 'dev.flowingspdg.vmix.function' && 'NOT YET!' }
    </>
  )
}

export default App
