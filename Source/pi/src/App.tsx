import { useEffect, useState } from 'react'
import { Preview, type PreviewSettings } from './components/preview'
import { Program, type ProgramSettings } from './components/program'
import type {
  SendToPropertyInspector,
  SendInputs,
  DestinationToInputs,
  ActionInfo,
} from './types/streamdeck'
import { Activator, type ActivatorSettings } from './components/activator'
import { headlessStreamDeck } from './adapters/stream-deck'

function App() {
  type T = PreviewSettings | ProgramSettings | ActivatorSettings
  // States
  const [settings, setSettings] = useState<T>({} as T)
  const [inputs, setInputs] = useState<DestinationToInputs>({})
  const [actionInfos, setActionInfos] = useState<ActionInfo<unknown>[]>(
    headlessStreamDeck.getInfos(),
  )

  useEffect(() => {
    const open = () => {
      console.log('Opened')
      setActionInfos(headlessStreamDeck.getInfos())
    }
    headlessStreamDeck.addEventListener('open', open)

    const didReceiveSettings = (s: unknown) => {
      console.log('Settings received', s)
      setSettings(s as T)
    }
    headlessStreamDeck.addEventListener(
      'didReceiveGlobalSettings',
      didReceiveSettings,
    )

    const sendToPropertyInspector = (payload: unknown) => {
      if (!payload) return
      if (typeof payload !== 'object') return
      if (!('event' in payload)) return

      if (payload?.event === 'inputs') {
        const p: SendToPropertyInspector<SendInputs> = payload as SendToPropertyInspector<SendInputs>
        console.log('inputs', p.payload.inputs)
        setInputs(p.payload.inputs)
      }
    }

    headlessStreamDeck.addEventListener(
      'sendToPropertyInspector',
      sendToPropertyInspector,
    )

    return () => {
      headlessStreamDeck.removeEventListener('open', open)
      headlessStreamDeck.removeEventListener(
        'didReceiveGlobalSettings',
        didReceiveSettings,
      )
    }
  }, [])

  const onSettingsUpdate = (s: T) => {
    console.log('Updated. sending payload...', s)
    setSettings(s)
    headlessStreamDeck.setSettings(s)
  }

  return (
    <>
      {actionInfos
        .map(info => info.action)
        .includes('dev.flowingspdg.vmix.preview') && (
          <Preview
            inputs={inputs}
            settings={settings as PreviewSettings}
            onUpdate={onSettingsUpdate}
          />
      )}

      {actionInfos
        .map(info => info.action)
        .includes('dev.flowingspdg.vmix.program') && (
          <Program
            inputs={inputs}
            settings={settings as ProgramSettings}
            onUpdate={onSettingsUpdate}
          />
      )}
      {actionInfos
        .map(info => info.action)
        .includes('dev.flowingspdg.vmix.activator') && (
          <Activator
            inputs={inputs}
            settings={settings as ActivatorSettings}
            onUpdate={onSettingsUpdate}
          />
      )}
      {actionInfos
        .map(info => info.action)
        .includes('dev.flowingspdg.vmix') && 'NOT YET!'}
    </>
  )
}

export default App
