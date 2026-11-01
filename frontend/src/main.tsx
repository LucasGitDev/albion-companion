import React from 'react'
import { createRoot } from 'react-dom/client'
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
