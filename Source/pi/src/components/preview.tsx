import type { input } from '../types/streamdeck'
import type { SD } from '../sd'

export type PreviewSettings = {
  dest: string
  input: number
  mix: number | null
  tally: boolean
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
          onClick={() => {
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
          onClick={() => {
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
        <div className="sdpi-item-label">Tally</div>

        <div className="sdpi-item-value">
          <input
            id="tally"
            type="checkbox"
            className="sdProperty sdCheckbox"
            checked={props.settings.tally}
            onChange={(e) => {
              props.onUpdate({
                ...props.settings,
                tally: e.target.checked,
              })
            }}
          />
          {/* biome-ignore lint/a11y/noLabelWithoutControl: <explanation> */}
          <label htmlFor="tally" className="sdpi-item-label"><span /></label>

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
