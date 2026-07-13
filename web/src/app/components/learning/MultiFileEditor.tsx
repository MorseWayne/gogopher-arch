import { AlertTriangle, Check, Cloud, CloudOff, GitCompareArrows, LoaderCircle, RotateCcw, Save } from 'lucide-react'

import type { AttemptResponse, Task } from '../../../api/learning'
import { useAttemptExecution } from '../../hooks/useAttemptExecution'
import { useAttemptWorkspace } from '../../hooks/useAttemptWorkspace'
import { GoCodeEditor } from '../GoCodeEditor'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import { WorkspaceExplorer } from './WorkspaceExplorer'
import { ActionBar } from './ActionBar'
import { ExecutionPanel } from './ExecutionPanel'
import { SubmissionPanel } from './SubmissionPanel'

export function MultiFileEditor({
  attempt,
  task,
  onAttemptChange,
}: {
  attempt: AttemptResponse
  task: Task
  onAttemptChange: (attempt: AttemptResponse) => void
}) {
  const workspace = useAttemptWorkspace(attempt, task, onAttemptChange)
  const execution = useAttemptExecution(attempt, {
    revision: workspace.baseRevision,
    hash: workspace.baseHash,
    dirty: workspace.dirty,
  }, onAttemptChange)
  const editable = task.editable_paths.includes(workspace.selectedPath)
  const selectedContents = workspace.files[workspace.selectedPath] ?? ''

  return (
    <div className="overflow-hidden rounded-2xl border bg-background">
      <div className="flex flex-wrap items-center gap-2 border-b px-4 py-3">
        <div className="mr-auto">
          <div className="text-sm font-semibold">多文件 workspace</div>
          <div className="mt-0.5 font-mono text-xs text-muted-foreground">revision {workspace.baseRevision} · {workspace.baseHash.slice(0, 12)}</div>
        </div>
        {workspace.recovered && <Badge variant="outline"><RotateCcw />已恢复未同步备份</Badge>}
        {workspace.dirty ? <Badge variant="secondary"><CloudOff />未同步</Badge> : <Badge variant="outline"><Cloud /><Check />已保存</Badge>}
        <Button onClick={() => void workspace.save()} disabled={!workspace.dirty || workspace.saveState.status === 'saving'}>
          {workspace.saveState.status === 'saving' ? <LoaderCircle className="animate-spin" /> : <Save />}
          保存到服务端
        </Button>
      </div>

      {workspace.saveState.status === 'error' && (
        <div role="alert" className="flex items-start gap-2 border-b border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          <AlertTriangle className="mt-0.5 size-4" />{workspace.saveState.message}
        </div>
      )}
      {workspace.saveState.status === 'conflict' && (
        <div role="alert" className="border-b border-amber-500/30 bg-amber-500/10 p-4">
          <div className="flex items-center gap-2 font-semibold"><GitCompareArrows className="size-4" />检测到另一标签页已保存</div>
          <div className="mt-3 grid gap-3 text-xs sm:grid-cols-2">
            <VersionCard label="本地未同步版本" revision={workspace.baseRevision} files={workspace.saveState.value.localFiles} />
            <VersionCard label="服务端当前版本" revision={workspace.saveState.value.serverAttempt.workspace_revision} files={workspace.saveState.value.serverAttempt.workspace} />
          </div>
          <p className="mt-3 text-xs text-muted-foreground">不会自动 merge。选择服务端会丢弃本地改动；选择本地会以最新 revision 继续编辑，仍需再次保存。</p>
          <div className="mt-3 flex flex-wrap gap-2">
            <Button variant="outline" size="sm" onClick={() => workspace.resolveConflict('server')}>重新载入服务端</Button>
            <Button size="sm" onClick={() => workspace.resolveConflict('local')}>保留本地并继续编辑</Button>
          </div>
        </div>
      )}

      <div className="grid min-h-[34rem] grid-cols-1 md:grid-cols-[190px_minmax(0,1fr)]">
        <WorkspaceExplorer
          paths={Object.keys(workspace.files).sort()}
          editablePaths={task.editable_paths}
          selectedPath={workspace.selectedPath}
          onSelect={workspace.setSelectedPath}
        />
        <div className="min-w-0 bg-[#0d1117]">
          <div className="flex h-10 items-center gap-2 border-b border-slate-800 px-4 font-mono text-xs text-slate-300">
            <span className="truncate">{workspace.selectedPath}</span>
            <span className="ml-auto text-slate-500">{editable ? 'editable' : 'readonly'}</span>
          </div>
          <GoCodeEditor
            key={workspace.selectedPath}
            ariaLabel={`${workspace.selectedPath} editor`}
            value={selectedContents}
            onChange={(value) => workspace.updateFile(workspace.selectedPath, value)}
            readOnly={!editable}
            syntax={workspace.selectedPath.endsWith('.go') ? 'go' : 'plain'}
            height="31.5rem"
          />
        </div>
      </div>
      <ActionBar
        allowedActions={task.allowed_actions}
        disabled={workspace.dirty || attempt.status !== 'active'}
        disabledMessage={attempt.status !== 'active'
          ? 'Attempt 已提交，工具动作已关闭。'
          : workspace.dirty ? '请先把当前 workspace 保存到服务端，再运行工具动作。' : undefined}
        busy={execution.commandState.status === 'sending' || execution.pollingExecutionID !== null}
        error={execution.commandState.status === 'error' ? execution.commandState.message : execution.pollingFailure?.message}
        retryLabel={execution.commandState.status === 'error' ? undefined : '重新读取 Execution 状态'}
        onRun={execution.run}
        onRetry={execution.commandState.status === 'error' ? execution.retry : execution.retryPolling}
      />
      <SubmissionPanel
        attempt={attempt}
        task={task}
        workspace={{
          revision: workspace.baseRevision,
          hash: workspace.baseHash,
          dirty: workspace.dirty,
          save: workspace.save,
        }}
        onAttemptChange={onAttemptChange}
      />
      <ExecutionPanel executions={attempt.executions} ruleResults={attempt.rule_results} />
    </div>
  )
}

function VersionCard({ label, revision, files }: { label: string; revision: number; files: Record<string, string> }) {
  return (
    <div className="rounded-lg border bg-background/80 p-3">
      <div className="font-semibold">{label}</div>
      <div className="mt-1 font-mono text-muted-foreground">revision {revision} · {Object.keys(files).length} files</div>
    </div>
  )
}
