import type { DestinationToInputs } from '../types/streamdeck'

export type ProgramSettings = {
  dest: string
  input: number
  mix: number | null
  tally: boolean
  transition: string
}

type ProgramProps = {
  settings: ProgramSettings
  inputs: DestinationToInputs

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
          value={props.settings.dest}
          onChange={
          e => props.onUpdate({
            ...props.settings,
            dest: e.target.value,
          })
        }
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
                input: Number.parseInt(e.target.value),
              })
            }}
          >

            {props.inputs[props.settings.dest]?.map(input => (
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