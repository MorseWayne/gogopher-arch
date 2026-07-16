import { Blocks, FlaskConical, LoaderCircle, RefreshCw, ShieldCheck } from 'lucide-react'

import type { ExecutionAction } from '../../../api/learning'
import { Button } from '../ui/button'

type RunnableAction = Exclude<ExecutionAction, 'submit'>

const actionMeta: Record<RunnableAction, { label: string; icon: typeof Blocks }> = {
  build: { label: 'Build', icon: Blocks },
  test: { label: 'Test', icon: FlaskConical },
  vet: { label: 'Vet', icon: ShieldCheck },
}

export function ActionBar({
  allowedActions,
  disabled,
  disabledMessage,
  busy,
  error,
  retryLabel,
  onRun,
  onRetry,
}: {
  allowedActions: ExecutionAction[]
  disabled: boolean
  disabledMessage?: string
  busy: boolean
  error?: string
  retryLabel?: string
  onRun: (action: RunnableAction) => void
  onRetry: () => void
}) {
  const actions = (['build', 'test', 'vet'] as RunnableAction[]).filter((action) => allowedActions.includes(action))
  return (
    <div className="border-t bg-muted/15 p-4">
      <div className="flex flex-wrap items-center gap-2">
        <div className="mr-auto">
          <div className="text-sm font-semibold">运行工具检查</div>
          <div className="text-xs text-muted-foreground">建议按 Build → Test → Vet 的顺序运行，并阅读每一步反馈。</div>
        </div>
        {actions.map((action) => {
          const Icon = actionMeta[action].icon
          return <Button key={action} variant="outline" onClick={() => onRun(action)} disabled={disabled || busy}>
            {busy ? <LoaderCircle className="animate-spin" /> : <Icon />}{actionMeta[action].label}
          </Button>
        })}
      </div>
      {disabled && disabledMessage && <p className="mt-2 text-xs text-amber-600">{disabledMessage}</p>}
      {error && <div role="alert" className="mt-3 flex items-center justify-between gap-3 rounded-lg bg-destructive/10 p-3 text-xs text-destructive"><span>{error}</span><Button size="sm" variant="outline" onClick={onRetry}><RefreshCw />{retryLabel ?? '重试这一步'}</Button></div>}
    </div>
  )
}
