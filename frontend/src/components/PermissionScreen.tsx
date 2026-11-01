import { ShieldAlert, RefreshCw, Terminal, Download, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'

type OS = 'mac' | 'windows' | 'linux'

function detectOS(): OS {
  const ua = navigator.userAgent
  if (/Mac/i.test(ua)) return 'mac'
  if (/Win/i.test(ua)) return 'windows'
  return 'linux'
}

const guides: Record<OS, { title: string; steps: { icon: typeof Check; text: string; code?: string }[] }> = {
  mac: {
    title: 'macOS needs read access to /dev/bpf*',
    steps: [
      { icon: Download, text: 'Install Wireshark — it ships the ChmodBPF helper that grants access at login. Then capture just works.' },
      { icon: ShieldAlert, text: 'Or install the one-time helper below (asks for your admin password once) and log back in.' },
      { icon: Terminal, text: 'Or, for a quick test, launch the CLI with sudo:', code: 'sudo ./companion-cli' },
    ],
  },
  linux: {
    title: 'Linux needs raw-capture capabilities',
    steps: [
      { icon: Terminal, text: 'Grant the binary capture capabilities (no root needed afterwards):', code: 'sudo setcap cap_net_raw,cap_net_admin+eip ./albion-companion' },
      { icon: ShieldAlert, text: 'Then restart the app and try again.' },
    ],
  },
  windows: {
    title: 'Windows needs Npcap',
    steps: [
      { icon: Download, text: 'Install Npcap in WinPcap-compatible mode from npcap.com.' },
      { icon: ShieldAlert, text: 'Restart the app after installing.' },
    ],
  },
}

export function PermissionScreen({ error, onRetry }: { error?: string; onRetry: () => void }) {
  const os = detectOS()
  const guide = guides[os]

  return (
    <div className="grid min-h-screen place-items-center p-6">
      <div className="w-full max-w-lg">
        <div className="mb-6 flex flex-col items-center text-center">
          <div className="mb-4 grid size-14 place-items-center rounded-2xl bg-primary/12 ring-1 ring-primary/25">
            <ShieldAlert className="size-7 text-primary" />
          </div>
          <h1 className="text-xl font-semibold tracking-tight">Capture permission needed</h1>
          <p className="mt-2 max-w-sm text-sm text-muted-foreground">
            Albion Companion only reads network packets — it never sends anything or touches the game. To read them it
            needs capture access on this machine.
          </p>
        </div>

        <div className="rounded-xl border border-border bg-surface/60 p-5">
          <div className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">{guide.title}</div>
          <ol className="flex flex-col gap-4">
            {guide.steps.map((s, i) => (
              <li key={i} className="flex gap-3">
                <div className="mt-0.5 grid size-6 shrink-0 place-items-center rounded-full bg-elevated text-[11px] font-semibold tnum text-muted-foreground">
                  {i + 1}
                </div>
                <div className="flex-1">
                  <div className="flex items-start gap-2 text-sm">
                    <s.icon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                    <span>{s.text}</span>
                  </div>
                  {s.code && (
                    <code className="mt-2 block overflow-x-auto rounded-md border border-border bg-background px-3 py-2 font-mono text-xs text-foreground/90">
                      {s.code}
                    </code>
                  )}
                </div>
              </li>
            ))}
          </ol>
        </div>

        {error && (
          <div className="mt-4 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive">
            {error}
          </div>
        )}

        <div className="mt-5 flex items-center justify-center gap-3">
          <Button onClick={onRetry}>
            <RefreshCw /> Try again
          </Button>
        </div>
        <p className="mt-4 text-center text-xs text-muted-foreground/70">
          Behind a VPN, ExitLag or GeForce Now? The tunnel hides game traffic from capture — disable it first.
        </p>
      </div>
    </div>
  )
}
