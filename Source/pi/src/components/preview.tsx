import type { DestinationToInputs } from '../types/streamdeck'
import type { SD } from '../sd'
import { TallyMode } from './tally'

export type PreviewSettings = {
  dest: string
  input: number
  mix: number
  tally_mode: TallyMode
}

export type PreviewProps = {
  settings: PreviewSettings
  inputs: DestinationToInputs
  destinations: string[]
  sd: SD<unknown> | null

  // Callback
  onUpdate: (settings: PreviewSettings) => void
}

export const Preview = (props: PreviewProps) => {
  console.log('received props', props)
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
        {props.destinations.includes(props.settings.dest) ? (
          <div className="sdpi-item-label">Connected</div>
        ) : (
          <>
            <div className="sdpi-item-label">Connection</div>
            <button
              type="button"
              className="sdpi-item-value"
              disabled={props.destinations.includes(props.settings.dest)}
              onClick={(e) => {
                e.preventDefault()
                props.sd?.sendValueToPlugin({
                  event: "connect",
                  payload: {
                    host: props.settings.dest,
                  },
                })
              }}
            >
              Connect
            </button>
          </>
        )}

        {props.destinations.includes(props.settings.dest) && (
          <button
            type="button"
            className="sdpi-item-value"
            disabled={!props.destinations.includes(props.settings.dest)}
            onClick={(e) => {
              e.preventDefault()
              props.sd?.sendValueToPlugin({
                event: "disconnect",
                payload: {
                  host: props.settings.dest,
                },
              })
            }}
          >
            Disconnect
          </button>
        )}
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
        <div className="sdpi-item-label">Mix</div>
        <div className="sdpi-item-value">
          <select
            className="sdProperty sdList"
            value={props.settings.mix ?? 0}
            onChange={(e) => {
              const mixValue = Number.parseInt(e.target.value, 10);
              props.onUpdate({
                ...props.settings,
                mix: mixValue,
                tally_mode: mixValue !== 0 ? TallyMode.ACTS : props.settings.tally_mode,
              })
            }}
          >
            {Array.from({length: 16}, (_, i) => {
              return (
                <option key={String(i)} value={i} selected={props.settings.mix === i}>
                  Mix{i+1}{i === 0 ? ' (Main)' : ''}
                </option>
              );
            })}
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
            <option 
              selected={props.settings.tally_mode === TallyMode.TALLY}
              key={TallyMode.TALLY}
              value={TallyMode.TALLY} 
              disabled={props.settings.mix !== 0}
            >
              TALLY
            </option>
            <option 
              selected={props.settings.tally_mode === TallyMode.ACTS} 
              key={TallyMode.ACTS} 
              value={TallyMode.ACTS}
            >
              ACTS
            </option>
            <option 
              selected={props.settings.tally_mode === TallyMode.DISABLED} 
              key={TallyMode.DISABLED} 
              value={TallyMode.DISABLED}
            >
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
            {(props.inputs[props.settings.dest] ?? []).map(input => (
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
