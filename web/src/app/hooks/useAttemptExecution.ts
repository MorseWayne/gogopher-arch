import { useCallback, useEffect, useRef, useState } from 'react'

import { executeAttempt, getAttempt, LearningApiError } from '../../api/learning'
import type { AttemptResponse, ExecuteAttemptRequest, Execution, ExecutionAction } from '../../api/learning'

type RunnableAction = Exclude<ExecutionAction, 'submit'>

interface PendingCommand {
  action: RunnableAction
  request: ExecuteAttemptRequest
}

type CommandState =
  | { status: 'idle' | 'sending' }
  | { status: 'error'; message: string; pending: PendingCommand }

const terminalStatuses = new Set(['succeeded', 'user_failed', 'infra_failed'])

export function useAttemptExecution(
  attempt: AttemptResponse,
  workspace: { revision: number; hash: string; dirty: boolean },
  onAttemptChange: (attempt: AttemptResponse) => void,
  pollIntervalMS = 750,
) {
  const [commandState, setCommandState] = useState<CommandState>({ status: 'idle' })
  const [pollingExecutionID, setPollingExecutionID] = useState<string | null>(null)
  const [pollingFailure, setPollingFailure] = useState<{ executionID: string; message: string } | null>(null)
  const pollingRef = useRef<string | null>(null)
  const mountedRef = useRef(true)

  useEffect(() => {
    mountedRef.current = true
    return () => {
      mountedRef.current = false
    }
  }, [])

  const poll = useCallback(async (executionID: string) => {
    if (pollingRef.current) return
    pollingRef.current = executionID
    setPollingExecutionID(executionID)
    setPollingFailure(null)
    try {
      for (let count = 0; count < 120 && mountedRef.current; count += 1) {
        if (pollIntervalMS > 0) await delay(pollIntervalMS)
        const refreshed = await getAttempt(attempt.id)
        if (!mountedRef.current) return
        onAttemptChange(refreshed)
        const execution = refreshed.executions.find((item) => item.id === executionID)
        if (!execution || terminalStatuses.has(execution.status)) return
      }
    } catch (error) {
      if (mountedRef.current) setPollingFailure({ executionID, message: errorText(error) })
    } finally {
      pollingRef.current = null
      if (mountedRef.current) setPollingExecutionID(null)
    }
  }, [attempt.id, onAttemptChange, pollIntervalMS])

  useEffect(() => {
    const pending = [...attempt.executions].reverse().find((execution) =>
      execution.status === 'queued' || execution.status === 'running')
    if (pending) void poll(pending.id)
  }, [attempt.executions, poll])

  const send = useCallback(async (pending: PendingCommand) => {
    setCommandState({ status: 'sending' })
    try {
      const execution = await executeAttempt(attempt.id, pending.request)
      setCommandState({ status: 'idle' })
      onAttemptChange(mergeExecution(attempt, execution))
    } catch (error) {
      if (error instanceof LearningApiError && error.code === 'idempotency_conflict' && error.payload.execution_id) {
        try {
          const refreshed = await getAttempt(attempt.id)
          onAttemptChange(refreshed)
          setCommandState({ status: 'idle' })
          return
        } catch (refreshError) {
          setCommandState({ status: 'error', message: errorText(refreshError), pending })
          return
        }
      }
      setCommandState({ status: 'error', message: errorText(error), pending })
    }
  }, [attempt, onAttemptChange])

  const run = useCallback((action: RunnableAction) => {
    if (workspace.dirty || commandState.status === 'sending' || pollingExecutionID) return
    void send({
      action,
      request: {
        request_key: requestKey(),
        action,
        workspace_revision: workspace.revision,
        workspace_hash: workspace.hash,
      },
    })
  }, [commandState.status, pollingExecutionID, send, workspace])

  const retry = useCallback(() => {
    if (commandState.status === 'error' && commandState.pending.request.request_key) {
      void send(commandState.pending)
    }
  }, [commandState, send])

  const retryPolling = useCallback(() => {
    if (pollingFailure) void poll(pollingFailure.executionID)
  }, [poll, pollingFailure])

  return { commandState, pollingExecutionID, pollingFailure, retry, retryPolling, run }
}

function mergeExecution(attempt: AttemptResponse, execution: Execution): AttemptResponse {
  const executions = attempt.executions.filter((item) => item.id !== execution.id)
  return { ...attempt, executions: [...executions, execution] }
}

function requestKey(): string {
  return globalThis.crypto?.randomUUID?.() ?? `request-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function delay(milliseconds: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, milliseconds))
}

function errorText(error: unknown): string {
  if (error instanceof LearningApiError) return `${error.code}（HTTP ${error.status}）：${error.message}`
  if (error instanceof Error) return error.message
  return 'Execution 请求失败'
}
