import React from 'react'
import { createRoot } from 'react-dom/client'
// Fonts are bundled by Vite (offline-safe for the Wails desktop app, no CDN).
import '@fontsource-variable/space-grotesk' // display — titles / big numbers
import '@fontsource-variable/geist' // body / UI
import '@fontsource-variable/jetbrains-mono' // mono
import './style.css'
import App from './App'

// In a plain browser (vite dev, no Wails runtime) install the mock bridge so the
// UI can be developed and reviewed without the Go backend.
if (import.meta.env.DEV && typeof window.go === 'undefined') {
  const { installMock } = await import('./dev/mock')
  installMock()
}

const container = document.getElementById('root')
const root = createRoot(container!)
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
