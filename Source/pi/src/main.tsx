import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import { headlessStreamDeck } from './adapters/stream-deck'

import './index.css'

window.connectElgatoStreamDeckSocket = (
  inPort,
  inUUID,
  inRegisterEvent,
  inInfo,
  inActionInfo,
) => {
  headlessStreamDeck.add(inPort, {
    inPropertyInspectorUUID: inUUID,
    inRegisterEvent,
    inInfo,
    inActionInfo,
  })
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
