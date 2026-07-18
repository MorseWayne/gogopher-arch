import { useState, type FormEvent } from 'react'
import { FileCode2, FilePlus2, FileText, LockKeyhole, Trash2, X } from 'lucide-react'

import { ScrollArea } from '../ui/scroll-area'
import { Button } from '../ui/button'
import { Input } from '../ui/input'

export function WorkspaceExplorer({
  paths,
  isEditable,
  selectedPath,
  onSelect,
  canCreate,
  canDelete,
  onCreate,
  onDelete,
}: {
  paths: string[]
  isEditable: (path: string) => boolean
  selectedPath: string
  onSelect: (path: string) => void
  canCreate: boolean
  canDelete: boolean
  onCreate: (path: string) => string | null
  onDelete: (path: string) => void
}) {
  const [creating, setCreating] = useState(false)
  const [path, setPath] = useState('')
  const [error, setError] = useState('')

  function submit(event: FormEvent) {
    event.preventDefault()
    const result = onCreate(path)
    if (result) {
      setError(result)
      return
    }
    setPath('')
    setError('')
    setCreating(false)
  }

  return (
    <div className="h-full border-b bg-muted/20 md:border-r md:border-b-0">
      <div className="flex items-center gap-2 border-b px-3 py-2.5 text-xs font-semibold tracking-[0.14em] text-muted-foreground">
        <span className="mr-auto">练习文件</span>
        {canCreate && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={creating ? '取消新建文件' : '新建文件'}
            onClick={() => {
              setCreating((value) => !value)
              setError('')
            }}
          >
            {creating ? <X /> : <FilePlus2 />}
          </Button>
        )}
      </div>
      {creating && (
        <form className="space-y-2 border-b p-2" onSubmit={submit}>
          <Input
            autoFocus
            value={path}
            aria-label="新文件路径"
            placeholder="cmd/tool/main.go"
            onChange={(event) => setPath(event.target.value)}
          />
          {error && <p role="alert" className="text-[11px] leading-4 text-destructive">{error}</p>}
          <Button type="submit" size="sm" className="w-full">创建空文件</Button>
        </form>
      )}
      <ScrollArea className="h-48 md:h-[33rem]">
        <div className="space-y-1 p-2">
          {paths.map((path) => {
            const editable = isEditable(path)
            return (
              <div key={path} className={`group flex items-center rounded-lg transition-colors ${selectedPath === path ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'}`}>
                <button
                  type="button"
                  onClick={() => onSelect(path)}
                  className="flex min-w-0 flex-1 items-center gap-2 px-2.5 py-2 text-left font-mono text-xs"
                >
                  {path.endsWith('.md') ? <FileText className="size-3.5" /> : <FileCode2 className="size-3.5" />}
                  <span className="min-w-0 flex-1 truncate">{path}</span>
                  {!editable && <LockKeyhole aria-label="只读" className="size-3" />}
                </button>
                {canDelete && editable && (
                  <button
                    type="button"
                    aria-label={`删除 ${path}`}
                    onClick={() => onDelete(path)}
                    className="mr-1 rounded p-1 opacity-70 hover:bg-background/20 hover:opacity-100"
                  >
                    <Trash2 className="size-3" />
                  </button>
                )}
              </div>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}
