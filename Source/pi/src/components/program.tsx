import type { input } from '../types/vmix'

export type ProgramSettings = {
  host: string
  port: number
  input: string
  tally: boolean
  mix: number
  transition: string
}

type ProgramProps = {
  settings: ProgramSettings
  inputs: input[]

  // Callback
  onUpdate: (settings: ProgramSettings) => void
}

export const Program = (props: ProgramProps) => {
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
        <div className="sdpi-item-label">Port</div>
        <input
          className="sdpi-item-value"
          value={props.settings.port}
          onChange={(e) => {
            const port = Number.parseInt(e.target.value)
            if (!Number.isNaN(port)) {
              props.onUpdate({
                ...props.settings,
                port: port,
              })
            }
          }}
        />
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Mix</div>
        <input
          className="sdpi-item-value"
          value={props.settings.mix}
          onChange={(e) => {
            const mix = Number.parseInt(e.target.value)
            if (!Number.isNaN(mix)) {
              props.onUpdate({
                ...props.settings,
                mix: mix,
              })
            }
          }}
        />
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Transition</div>
        <input
          className="sdpi-item-value"
          value={props.settings.transition}
          onChange={
          e => props.onUpdate({
            ...props.settings,
            transition: e.target.value,
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
                input: e.target.value,
              })
            }}
          >

            {props.inputs.map((input) => {
              return (
                <option key={input.key} value={input.key}>
                  {input.number}
                  :
                  {' '}
                  {input.name}
                </option>
              )
            })}

          </select>
        </div>
      </div>

    </div>
  )
}
