export function Placeholder({ title, note }: { title: string; note?: string }) {
  return (
    <div className="flex h-full flex-col items-center justify-center text-center text-muted-foreground">
      <div className="text-lg font-medium text-foreground">{title}</div>
      {note && <div className="mt-1 max-w-sm text-sm">{note}</div>}
    </div>
  )
}
