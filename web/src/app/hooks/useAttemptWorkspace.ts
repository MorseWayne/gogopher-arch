import { useCallback, useEffect, useMemo, useState } from 'react'

import { getAttempt, LearningApiError, saveWorkspace } from '../../api/learning'
import type { AttemptResponse, Task } from '../../api/learning'

const BACKUP_VERSION = 1

interface DraftBackup {
  version: typeof BACKUP_VERSION
  attempt_id: string
  base_revision: number
  base_hash: string
  files: Record<string, string>
  saved_at: string
}

export interface WorkspaceConflict {
  localFiles: Record<string, string>
  serverAttempt: AttemptResponse
}

type SaveState =
  | { status: 'idle' | 'saving' | 'saved' }
  | { status: 'error'; message: string }
  | { status: 'conflict'; value: WorkspaceConflict }

export function draftBackupKey(attemptID: string, revision: number): string {
  return `gogopher:learning:draft:${attemptID}:${revision}`
}

export function useAttemptWorkspace(
  initialAttempt: AttemptResponse,
  task: Task,
  onAttemptChange: (attempt: AttemptResponse) => void,
) {
  const initial = useMemo(() => restoreDraft(initialAttempt, task), [initialAttempt.id])
  const [files, setFiles] = useState(initial.files)
  const [baseRevision, setBaseRevision] = useState(initialAttempt.workspace_revision)
  const [baseHash, setBaseHash] = useState(initialAttempt.workspace_hash)
  const [dirty, setDirty] = useState(initial.recovered)
  const [recovered, setRecovered] = useState(initial.recovered)
  const [selectedPath, setSelectedPath] = useState(
    task.editable_paths.find((path) => path in initial.files) ?? Object.keys(initial.files).sort()[0] ?? '',
  )
  const [saveState, setSaveState] = useState<SaveState>({ status: 'idle' })

  useEffect(() => {
    const key = draftBackupKey(initialAttempt.id, baseRevision)
    if (!dirty) {
      localStorage.removeItem(key)
      return
    }
    const backup: DraftBackup = {
      version: BACKUP_VERSION,
      attempt_id: initialAttempt.id,
      base_revision: baseRevision,
      base_hash: baseHash,
      files,
      saved_at: new Date().toISOString(),
    }
    localStorage.setItem(key, JSON.stringify(backup))
  }, [baseHash, baseRevision, dirty, files, initialAttempt.id])

  const updateFile = useCallback((path: string, contents: string) => {
    if (!task.editable_paths.includes(path)) return
    setFiles((current) => ({ ...current, [path]: contents }))
    setDirty(true)
    setRecovered(false)
    setSaveState({ status: 'idle' })
  }, [task.editable_paths])

  const save = useCallback(async () => {
    const validation = validateWorkspace(files, task)
    if (validation) {
      setSaveState({ status: 'error', message: validation })
      return null
    }
    setSaveState({ status: 'saving' })
    const staleKey = draftBackupKey(initialAttempt.id, baseRevision)
    try {
      const saved = await saveWorkspace(initialAttempt.id, { base_revision: baseRevision, files })
      localStorage.removeItem(staleKey)
      setFiles(saved.workspace)
      setBaseRevision(saved.workspace_revision)
      setBaseHash(saved.workspace_hash)
      setDirty(false)
      setRecovered(false)
      setSaveState({ status: 'saved' })
      onAttemptChange(saved)
      return saved
    } catch (error) {
      if (error instanceof LearningApiError && error.status === 409 &&
        (error.code === 'revision_conflict' || error.code === 'workspace_conflict')) {
        try {
          const serverAttempt = await getAttempt(initialAttempt.id)
          setSaveState({ status: 'conflict', value: { localFiles: { ...files }, serverAttempt } })
          return null
        } catch (refreshError) {
          setSaveState({ status: 'error', message: errorText(refreshError) })
          return null
        }
      }
      setSaveState({ status: 'error', message: errorText(error) })
      return null
    }
  }, [baseRevision, files, initialAttempt.id, onAttemptChange, task])

  const resolveConflict = useCallback((choice: 'server' | 'local') => {
    if (saveState.status !== 'conflict') return
    const { localFiles, serverAttempt } = saveState.value
    localStorage.removeItem(draftBackupKey(initialAttempt.id, baseRevision))
    setBaseRevision(serverAttempt.workspace_revision)
    setBaseHash(serverAttempt.workspace_hash)
    onAttemptChange(serverAttempt)
    if (choice === 'server') {
      setFiles(serverAttempt.workspace)
      setDirty(false)
    } else {
      setFiles(localFiles)
      setDirty(true)
    }
    setRecovered(false)
    setSaveState({ status: 'idle' })
  }, [baseRevision, initialAttempt.id, onAttemptChange, saveState])

  return {
    baseHash,
    baseRevision,
    dirty,
    files,
    recovered,
    resolveConflict,
    save,
    saveState,
    selectedPath,
    setSelectedPath,
    updateFile,
  }
}

function restoreDraft(attempt: AttemptResponse, task: Task): { files: Record<string, string>; recovered: boolean } {
  const raw = localStorage.getItem(draftBackupKey(attempt.id, attempt.workspace_revision))
  if (!raw) return { files: attempt.workspace, recovered: false }
  try {
    const value = JSON.parse(raw) as Partial<DraftBackup>
    if (value.version !== BACKUP_VERSION || value.attempt_id !== attempt.id ||
      value.base_revision !== attempt.workspace_revision || value.base_hash !== attempt.workspace_hash ||
      !isValidBackupFiles(value.files, attempt.workspace, task.editable_paths)) {
      localStorage.removeItem(draftBackupKey(attempt.id, attempt.workspace_revision))
      return { files: attempt.workspace, recovered: false }
    }
    return { files: value.files, recovered: true }
  } catch {
    localStorage.removeItem(draftBackupKey(attempt.id, attempt.workspace_revision))
    return { files: attempt.workspace, recovered: false }
  }
}

function isValidBackupFiles(
  value: unknown,
  serverFiles: Record<string, string>,
  editablePaths: string[],
): value is Record<string, string> {
  if (typeof value !== 'object' || value === null) return false
  const entries = Object.entries(value)
  if (entries.length !== Object.keys(serverFiles).length) return false
  return entries.every(([path, contents]) =>
    path in serverFiles && typeof contents === 'string' &&
    (editablePaths.includes(path) || contents === serverFiles[path]),
  )
}

function validateWorkspace(files: Record<string, string>, task: Task): string | null {
  const entries = Object.entries(files)
  if (entries.length > task.limits.max_files) {
    return `文件数量 ${entries.length} 超过限制 ${task.limits.max_files}`
  }
  let total = 0
  for (const [path, contents] of entries) {
    const size = new TextEncoder().encode(contents).length
    if (size > task.limits.max_file_bytes) {
      return `${path} 为 ${size} bytes，超过单文件限制 ${task.limits.max_file_bytes} bytes`
    }
    total += size
  }
  if (total > task.limits.max_total_bytes) {
    return `workspace 为 ${total} bytes，超过总限制 ${task.limits.max_total_bytes} bytes`
  }
  return null
}

function errorText(error: unknown): string {
  if (error instanceof LearningApiError) return `${error.code}（HTTP ${error.status}）：${error.message}`
  if (error instanceof Error) return error.message
  return 'workspace 保存失败'
}
