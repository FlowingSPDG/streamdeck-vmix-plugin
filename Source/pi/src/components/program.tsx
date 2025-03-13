import type { DestinationToInputs, DestinationStatus } from '../types/streamdeck'
import type { SD } from '../sd'
import { TallyMode } from './tally'

export type ProgramSettings = {
  dest: string
  input: number
  mix: number
  tally_mode: TallyMode
  transition: string
  duration: number
}

export type ProgramProps = {
  settings: ProgramSettings
  inputs: DestinationToInputs
  destinations: DestinationStatus[]
  sd: SD<unknown> | null

  // Callback
  onUpdate: (settings: ProgramSettings) => void
}

export const Program = (props: ProgramProps) => {
  console.log('received props', props)
  const currentDest = props.destinations.find(d => d.address === props.settings.dest)
  const isConnected = currentDest?.connected ?? false

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
        {isConnected ? (
          <div className="sdpi-item-label">Connected</div>
        ) : (
          <>
            <div className="sdpi-item-label">Connection</div>
            <button
              type="button"
              className="sdpi-item-value"
              disabled={isConnected}
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

        {isConnected && (
          <button
            type="button"
            className="sdpi-item-value"
            disabled={isConnected}
          >
            Already Connected!
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
              <option key={dest.address} value={dest.address}>
                {dest.address} {dest.connected ? '(Connected)' : '(Disconnected)'}
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
                <option key={String(i)} value={i}>
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
              key={TallyMode.TALLY}
              value={TallyMode.TALLY}
              disabled={props.settings.mix !== 0}
            >
              TALLY
            </option>
            <option
              key={TallyMode.ACTS}
              value={TallyMode.ACTS}
            >
              ACTS
            </option>
            <option
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

      <div className="sdpi-item">
        <div className="sdpi-item-label">Transition</div>
        <div className="sdpi-item-value">
          <select
            className="sdProperty sdList"
            id="transition"
            value={props.settings.transition}
            onChange={(e) => {
              props.onUpdate({
                ...props.settings,
                transition: e.target.value,
              })
            }}
          >
            <option value="ActiveInput">ActiveInput</option>
            <option value="Cut">Cut</option>
            <option value="CutDirect">CutDirect</option>
            <option value="Fade">Fade</option>
            <option value="Merge">Merge</option>
            <option value="Slide">Slide</option>
            <option value="Zoom">Zoom</option>
            <option value="Stinger1">Stinger1</option>
            <option value="Stinger2">Stinger2</option>
            <option value="Stinger3">Stinger3</option>
            <option value="Stinger4">Stinger4</option>
          </select>
        </div>
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Duration</div>
        <div className="sdpi-item-value">
          <input
            type="number"
            value={props.settings.duration}
            onChange={(e) => {
              props.onUpdate({
                ...props.settings,
                duration: Number.parseInt(e.target.value),
              })
            }}
          />
        </div>
      </div>
    </div>
  )
}
