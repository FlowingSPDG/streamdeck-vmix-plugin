import type { input } from '../types/vmix'

export type TallySettings = {
  host: string
  port: number
  input: string
  mix: number
  preview: boolean
  program: boolean
}

type TallyProps = {
  settings: TallySettings
  inputs: input[]

  // Callback
  onUpdate: (settings: TallySettings) => void
}

export const Tally = (props: TallyProps) => {
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
        <div className="sdpi-item-label">Preview Tally</div>

        <div className="sdpi-item-child">
            <input
              id="preview_tally"
              type="checkbox"
              className="sdProperty sdCheckbox"
              checked={props.settings.preview}
              onChange={(e) => {
                props.onUpdate({
                  ...props.settings,
                  preview: e.target.checked,
                })
              }}
            />
            <label htmlFor="preview_tally" className="sdpi-item-label"><span /></label>
        </div>
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Program Tally</div>

        <div className="sdpi-item-child">
            <input
              id="program_tally"
              type="checkbox"
              className="sdProperty sdCheckbox"
              checked={props.settings.program}
              onChange={(e) => {
                props.onUpdate({
                  ...props.settings,
                  program: e.target.checked,
                })
              }}
            />
            <label htmlFor="program_tally" className="sdpi-item-label"><span /></label>
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
