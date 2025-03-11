import type { input } from '../types/streamdeck'
import type { SD } from '../sd'

export type PreviewSettings = {
  dest: string
  input: number
  mix: number | null
  tally_mode: TallyMode
}

type TallyMode = 0 | 1 | 2 | 3

const tallyType = {
  UNKNOWN: 0 as const,
  TALLY: 1 as const,
  ACTS: 2 as const,
  DISABLED: 3 as const,
}

type PreviewProps = {
  settings: PreviewSettings
  inputs: input[]
  destinations: string[]
  sd: SD<unknown>

  // Callback
  onUpdate: (settings: PreviewSettings) => void
}

export const Preview = (props: PreviewProps) => {
  console.log('received props inputs', props.inputs)
  return (
    <div className="sdpi-wrapper">

      <div className="sdpi-item">
        <div className="sdpi-item-label">Connection</div>

        <input
          className="sdpi-item-value"
          value={props.settings.dest}
          onChange={(e) => {
            props.onUpdate({
              ...props.settings,
              dest: e.target.value,
            })
          }}
        />

      </div>

      <div className="sdpi-item">

        {/* biome-ignore lint/style/useSelfClosingElements: <explanation> */}
        <div className="sdpi-item-label"></div>
        <button
          type="button"
          className="sdpi-item-value"
          disabled={props.destinations.includes(props.settings.dest)}
          onClick={(e) => {
            e.preventDefault()
            props.sd.sendValueToPlugin({
              event: "connect",
              payload: {
                host: props.settings.dest,
              },
            })
          }}
        >
          Connect
        </button>

        <button
          type="button"
          className="sdpi-item-value"
          disabled={!props.destinations.includes(props.settings.dest)}
          onClick={(e) => {
            e.preventDefault()
            props.sd.sendValueToPlugin({
              event: "disconnect",
              payload: {
                host: props.settings.dest,
              },
            })
          }}
        >
          Disconnect
        </button>
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">vMix</div>
        <div className="sdpi-item-value">
          <select
            className="sdProperty sdList"
            id="host"
            value={props.settings.dest}
            onChange={(e) => {
              props.onUpdate({
                ...props.settings,
                dest: e.target.value,
              })
            }}
          >

            {props.destinations.map(dest => (
              <option key={dest} value={dest}>
                {dest}
              </option>
            ))}

          </select>
        </div>
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Tally Mode</div>
        <div className="sdpi-item-value">
          <select
            className="sdProperty sdList"
            id="tallyMode"
            value={props.settings.tally_mode as number}
            onChange={(e) => {
              props.onUpdate({
                ...props.settings,
                tally_mode: Number.parseInt(e.target.value) as TallyMode,
              })
            }}
          >

            <option key={tallyType.TALLY} value={tallyType.TALLY}>
              TALLY
            </option>

            <option key={tallyType.ACTS} value={tallyType.ACTS}>
              ACTS
            </option>

            <option key={tallyType.DISABLED} value={tallyType.DISABLED}>
              DISABLED
            </option>

          </select>
        </div>
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Input</div>
        <div className="sdpi-item-value">
          <select
            className="sdProperty sdList"
            id="inputs"
            value={props.settings.input}
            onChange={(e) => {
              props.onUpdate({
                ...props.settings,
                input: Number.parseInt(e.target.value),
              })
            }}
          >

            {(props.inputs ?? []).map(input => (
              <option key={input.key} value={input.number}>
                {input.number}
                {' '}
                [
                {input.name}
                ]
              </option>
            ))}

          </select>
        </div>
      </div>

    </div>
  )
}
