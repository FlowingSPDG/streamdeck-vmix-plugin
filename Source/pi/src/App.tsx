import { useCallback, useSyncExternalStore } from 'react'
import { Preview, type PreviewSettings } from './components/preview'
import { Program, type ProgramSettings } from './components/program'
import { Activator, type ActivatorSettings } from './components/activator'
import { headlessStreamDeck } from './adapters/stream-deck'
import { settingsStore, inputsStore, actionInfoStore } from './stores'

type T = PreviewSettings | ProgramSettings | ActivatorSettings

function App() {
  /**
   * useSyncExternalStore は getValue で受け入れた値が not shallow equal だと
   * 再レンダリングトリガーされるため getValue は cached な値である必要がある
   * headlessStreamDeck から返却される getActionInfo, getActionInfos は計算して返却されるため、object が毎回異なる
   * xxxStore は event があるたびに値を更新し、それを保持するため getValue は cached な値になるため Infinite Loop を回避する
   * ちなみにこの辺は rxjs と jotai を使うと簡単に回避できるのだが、jotai は内部的に useSyncExternalStore を使っていないため本質的には別物になる
   **/
  const inputs = useSyncExternalStore(
    inputsStore.subscribe,
    inputsStore.getValue,
  )
  const settings = useSyncExternalStore(
    settingsStore.subscribe,
    settingsStore.getValue,
  )
  const actionInfos = useSyncExternalStore(
    actionInfoStore.subscribe,
    actionInfoStore.getValue,
  )

  const onSettingsUpdate = useCallback((s: T) => {
    console.log('Updated. sending payload...', s)
    headlessStreamDeck.setSettings(s)
  }, [])

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
