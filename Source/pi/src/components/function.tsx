import type { SD } from '../sd'
import type { DestinationStatus } from '../types/streamdeck'

export interface FunctionSettings {
  dest: string
  function: string
  query: string
  input: number | null
  acts_event: string
  acts_input: string
  acts_active_state: string
  acts_inactive_state: string
}

interface FunctionProps {
  settings: FunctionSettings
  destinations: DestinationStatus[]
  onUpdate: (settings: FunctionSettings) => void
  sd: SD<unknown>
}

export function FunctionComponent(props: FunctionProps) {
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
            disabled={!isConnected}
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
              <option key={dest.address} value={dest.address}>
                {dest.address} {dest.connected ? '(Connected)' : '(Disconnected)'}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Function Name</div>
        <input
          className="sdpi-item-value"
          type="text"
          value={props.settings.function}
          onChange={(e) => props.onUpdate({ ...props.settings, function: e.target.value })}
          placeholder="Enter vMix function name"
        />
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Query</div>
        <input
          className="sdpi-item-value"
          type="text" 
          value={props.settings.query}
          onChange={(e) => props.onUpdate({ ...props.settings, query: e.target.value })}
          placeholder="Enter query parameters"
        />
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Acts Event</div>
        <input
          className="sdpi-item-value"
          type="text"
          value={props.settings.acts_event}
          onChange={(e) => props.onUpdate({ ...props.settings, acts_event: e.target.value })}
          placeholder="Enter acts target event. e.g. InputPreview"
        />
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Acts Input</div>
        <input
          className="sdpi-item-value"
          type="number"
          value={props.settings.acts_input}
          onChange={(e) => props.onUpdate({ ...props.settings, acts_input: e.target.value })}
          placeholder="Enter acts target input. e.g. 1."
        />
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Active State</div>
        <input
          className="sdpi-item-value"
          type="text"
          value={props.settings.acts_active_state}
          onChange={(e) => props.onUpdate({ ...props.settings, acts_active_state: e.target.value })}
          placeholder="Enter acts target state. e.g. 1. or leave blank."
        />
      </div>

      <div className="sdpi-item">
        <div className="sdpi-item-label">Inactive State</div>
        <input
          className="sdpi-item-value"
          type="text"
          value={props.settings.acts_inactive_state}
          onChange={(e) => props.onUpdate({ ...props.settings, acts_inactive_state: e.target.value })}
          placeholder="Enter acts inactive state. e.g. 0. or leave blank."
        />
      </div>
    </div>
  )
}