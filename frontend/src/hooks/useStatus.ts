import { useEffect, useState } from 'react'
import { api, on, type CaptureStatus } from '@/lib/bindings'

/**
 * Capture status with a live packet counter. The backend only emits
 * capture:status on start/stop, so a running capture's packet/message counts
 * would otherwise never advance. We poll once a second while mounted and also
 * refresh immediately on the capture:status event.
 */
export function useStatus(): CaptureStatus | null {
  const [status, setStatus] = useState<CaptureStatus | null>(null)

  useEffect(() => {
    let alive = true
    const refresh = () =>
      api.status().then((s) => {
        if (alive) setStatus(s)
      })
    refresh()
    const off = on('capture:status', refresh)
    const timer = setInterval(refresh, 1000)
    return () => {
      alive = false
      off()
      clearInterval(timer)
    }
  }, [])

  return status
}
