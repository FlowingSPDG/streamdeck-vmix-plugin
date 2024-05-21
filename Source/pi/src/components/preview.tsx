import type { input } from '../types/vmix'

export type PreviewSettings = {
  host: string
  input: number
  tally: boolean
}

type PreviewProps = {
  settings: PreviewSettings
  inputs: input[]

  // Callback
  onUpdate: (settings: PreviewSettings) => void
}

export const Preview = (props: PreviewProps) => {
  return (
    <div className="sdpi-wrapper">
      <div className="sdpi-item">
        <div className="sdpi-item-label">Host IP</div>
        <input
          className="sdpi-item-value"
          value={props.settings.host}
          onChange={
          e => props.onUpdate({
            ...props.settings,
            host: e.target.value,
          })
        }
        />
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Tally</div>

        <div className="sdpi-item-child">
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
            <label htmlFor="tally" className="sdpi-item-label"><span /></label>

        </div>
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Input</div>
        <div className="sdpi-item-child">
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

            {props.inputs.map((input) => {
              return (
                <option key={input.key} value={input.number}>
                  {input.number} [{input.name}]
                </option>
              )
            })}

          </select>
        </div>
      </div>

    </div>
  )
}
