import type { DestinationToInputs } from '../types/streamdeck'

export type ActivatorSettings = {
  dest: string
  input: number
  color: 1 | 2
  activator: 'Input' | 'InputPreview'
  | 'Overlay1' | 'Overlay2' | 'Overlay3' | 'Overlay4'
  | 'InputMix2' | 'InputMix3' | 'InputMix4'
  | 'InputPreviewMix2' | 'InputPreviewMix3' | 'InputPreviewMix4'
}

const checkActivator = (activator: string): activator is ActivatorSettings['activator'] => {
  return ['Input', 'InputPreview', 'Overlay1', 'Overlay2', 'Overlay3', 'Overlay4', 'InputMix2', 'InputMix3', 'InputMix4', 'InputPreviewMix2', 'InputPreviewMix3', 'InputPreviewMix4'].includes(activator)
}

type ActivatorProps = {
  settings: ActivatorSettings
  inputs: DestinationToInputs

  // Callback
  onUpdate: (settings: ActivatorSettings) => void
}

export const Activator = (props: ActivatorProps) => {
  console.log('props: ', props)
  console.log('inputs: ', props.inputs[props.settings.dest])
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
        <div className="sdpi-item-label">Tally Type</div>
        <div className="sdpi-item-child">
          <select
            className="sdProperty sdList"
            id="tally"
            value={props.settings.activator}
            onChange={(e) => {
              if (checkActivator(e.target.value)) {
                props.onUpdate({
                  ...props.settings,
                  activator: e.target.value,
                })
              }
            }}
          >

            <option value="InputPreview">PRV</option>
            <option value="Input">PGM</option>
            <option value="Overlay1">Overlay1</option>
            <option value="Overlay2">Overlay2</option>
            <option value="Overlay3">Overlay3</option>
            <option value="Overlay4">Overlay4</option>
            <option value="InputPreviewMix2">Mix2 PRV</option>
            <option value="InputMix2">Mix2 PGM</option>
            <option value="InputPreviewMix3">Mix3 PRV</option>
            <option value="InputMix3">Mix3 PGM</option>
            <option value="InputPreviewMix4">Mix4 PRV</option>
            <option value="InputMix4">Mix4 PGM</option>

          </select>
        </div>
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Tally Color</div>
        <div className="sdpi-item-child">
          <select
            className="sdProperty sdList"
            id="color"
            value={props.settings.color}
            onChange={(e) => {
              const value = Number.parseInt(e.target.value)
              if (value === 1 || value === 2) {
                props.onUpdate({
                  ...props.settings,
                  color: value,
                })
              }
            }}
          >

            <option value="1">Red</option>
            <option value="2">Green</option>

          </select>
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
