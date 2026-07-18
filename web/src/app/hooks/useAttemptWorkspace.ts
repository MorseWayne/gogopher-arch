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

  const isEditablePath = useCallback((path: string) =>
    task.editable_paths.includes(path) ||
    (task.workspace_policy.allow_new_files && !task.readonly_paths.includes(path)),
  [task.editable_paths, task.readonly_paths, task.workspace_policy.allow_new_files])

  const updateFile = useCallback((path: string, contents: string) => {
    if (!(path in files) || !isEditablePath(path)) return
    setFiles((current) => ({ ...current, [path]: contents }))
    setDirty(true)
    setRecovered(false)
    setSaveState({ status: 'idle' })
  }, [files, isEditablePath])

  const createFile = useCallback((rawPath: string): string | null => {
    const path = rawPath.trim()
    if (!task.workspace_policy.allow_new_files) return '本练习不允许新建文件'
    if (!validWorkspacePath(path)) return '请输入安全的相对路径，例如 cmd/tool/main.go'
    if (path in files) return '该文件已经存在'
    if (Object.keys(files).length >= task.limits.max_files) return `文件数量不能超过 ${task.limits.max_files}`
    setFiles((current) => ({ ...current, [path]: '' }))
    setSelectedPath(path)
    setDirty(true)
    setRecovered(false)
    setSaveState({ status: 'idle' })
    return null
  }, [files, task.limits.max_files, task.workspace_policy.allow_new_files])

  const deleteFile = useCallback((path: string) => {
    if (!task.workspace_policy.allow_delete_files || !isEditablePath(path) || !(path in files)) return
    const remaining = { ...files }
    delete remaining[path]
    setFiles(remaining)
    if (selectedPath === path) setSelectedPath(Object.keys(remaining).sort()[0] ?? '')
    setDirty(true)
    setRecovered(false)
    setSaveState({ status: 'idle' })
  }, [files, isEditablePath, selectedPath, task.workspace_policy.allow_delete_files])

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
    createFile,
    deleteFile,
    files,
    isEditablePath,
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
      !isValidBackupFiles(value.files, attempt.workspace, task)) {
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
  task: Task,
): value is Record<string, string> {
  if (typeof value !== 'object' || value === null) return false
  const candidate = value as Record<string, unknown>
  const entries = Object.entries(candidate)
  if (!task.workspace_policy.allow_delete_files && task.editable_paths.some((path) => !(path in candidate))) return false
  if (task.readonly_paths.some((path) => !(path in candidate) || candidate[path] !== serverFiles[path])) return false
  return entries.every(([path, contents]) => {
    if (typeof contents !== 'string' || !validWorkspacePath(path)) return false
    if (task.readonly_paths.includes(path)) return contents === serverFiles[path]
    return task.editable_paths.includes(path) || task.workspace_policy.allow_new_files
  }) && validateWorkspace(candidate as Record<string, string>, task) === null
}

function validateWorkspace(files: Record<string, string>, task: Task): string | null {
  const entries = Object.entries(files)
  if (entries.length > task.limits.max_files) {
    return `文件数量 ${entries.length} 超过限制 ${task.limits.max_files}`
  }
  let total = 0
  for (const [path, contents] of entries) {
    if (!validWorkspacePath(path)) return `文件路径 ${path} 不合法`
    const size = new TextEncoder().encode(contents).length
    if (size > task.limits.max_file_bytes) {
      return `${path} 为 ${size} 字节，超过单文件限制 ${task.limits.max_file_bytes} 字节`
    }
    total += size
  }
  if (total > task.limits.max_total_bytes) {
    return `全部文件共 ${total} 字节，超过总限制 ${task.limits.max_total_bytes} 字节`
  }
  return null
}

const workspacePathPattern = /^(?:[A-Za-z0-9._-]+\/)*[A-Za-z0-9._-]+$/

function validWorkspacePath(path: string): boolean {
  if (!workspacePathPattern.test(path) || path.startsWith('/') || path.includes('\\') || path.includes('\0')) return false
  return path.split('/').every((segment) => segment !== '' && segment !== '.' && segment !== '..')
}

function errorText(error: unknown): string {
  if (error instanceof LearningApiError) return '保存进度暂时失败，请重试。'
  if (error instanceof Error) return error.message
  return '保存进度暂时失败，请重试。'
}
