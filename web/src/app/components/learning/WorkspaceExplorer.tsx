import { FileCode2, FileText, LockKeyhole } from 'lucide-react'

import { ScrollArea } from '../ui/scroll-area'

export function WorkspaceExplorer({
  paths,
  editablePaths,
  selectedPath,
  onSelect,
}: {
  paths: string[]
  editablePaths: string[]
  selectedPath: string
  onSelect: (path: string) => void
}) {
  return (
    <div className="h-full border-b bg-muted/20 md:border-r md:border-b-0">
      <div className="border-b px-3 py-3 text-xs font-semibold tracking-[0.14em] text-muted-foreground">练习文件</div>
      <ScrollArea className="h-48 md:h-[33rem]">
        <div className="space-y-1 p-2">
          {paths.map((path) => {
            const editable = editablePaths.includes(path)
            return (
              <button
                key={path}
                type="button"
                onClick={() => onSelect(path)}
                className={`flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left font-mono text-xs transition-colors ${selectedPath === path ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'}`}
              >
                {path.endsWith('.md') ? <FileText className="size-3.5" /> : <FileCode2 className="size-3.5" />}
                <span className="min-w-0 flex-1 truncate">{path}</span>
                {!editable && <LockKeyhole aria-label="只读" className="size-3" />}
              </button>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}
