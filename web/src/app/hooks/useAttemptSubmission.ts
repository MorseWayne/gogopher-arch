import { useCallback, useRef, useState } from 'react'

import {
  getAttempt,
  LearningApiError,
  retrySubmission,
  submitAttempt,
} from '../../api/learning'
import type { AttemptResponse, SubmitAttemptRequest } from '../../api/learning'

interface SubmissionWorkspace {
  revision: number
  hash: string
  dirty: boolean
  save: () => Promise<AttemptResponse | null>
}

type PendingSubmission =
  | { kind: 'submit'; request: SubmitAttemptRequest }
  | { kind: 'retry'; submissionID: string; requestKey: string }

type SubmissionState =
  | { status: 'idle' | 'sending' }
  | { status: 'error'; message: string; pending?: PendingSubmission }

const recoverableConflictCodes = new Set([
  'attempt_already_submitted',
  'idempotency_conflict',
  'attempt_inactive',
])

export function useAttemptSubmission(
  attempt: AttemptResponse,
  workspace: SubmissionWorkspace,
  onAttemptChange: (attempt: AttemptResponse) => void,
) {
  const [state, setState] = useState<SubmissionState>({ status: 'idle' })
  const inFlight = useRef(false)

  const refresh = useCallback(async () => {
    const current = await getAttempt(attempt.id)
    onAttemptChange(current)
  }, [attempt.id, onAttemptChange])

  const perform = useCallback(async (pending: PendingSubmission) => {
    setState({ status: 'sending' })
    try {
      if (pending.kind === 'submit') {
        await submitAttempt(attempt.id, pending.request)
      } else {
        await retrySubmission(pending.submissionID, pending.requestKey)
      }
      await refresh()
      setState({ status: 'idle' })
    } catch (error) {
      if (error instanceof LearningApiError && error.status === 409 &&
        recoverableConflictCodes.has(error.code)) {
        try {
          await refresh()
          setState({ status: 'idle' })
          return
        } catch (refreshError) {
          setState({ status: 'error', message: errorText(refreshError), pending })
          return
        }
      }
      setState({ status: 'error', message: errorText(error), pending })
    } finally {
      inFlight.current = false
    }
  }, [attempt.id, refresh])

  const submit = useCallback(async (explanation = '') => {
    if (inFlight.current || attempt.status !== 'active') return
    inFlight.current = true
    setState({ status: 'sending' })

    let current = attempt
    if (workspace.dirty) {
      const saved = await workspace.save()
      if (!saved) {
        inFlight.current = false
        setState({ status: 'error', message: '代码尚未保存成功，本次没有提交。' })
        return
      }
      current = saved
    }

    const pending: PendingSubmission = {
      kind: 'submit',
      request: {
        submission_key: requestKey(),
        workspace_revision: current.workspace_revision,
        workspace_hash: current.workspace_hash,
        explanation,
      },
    }
    await perform(pending)
  }, [attempt, perform, workspace])

  const retryInfrastructure = useCallback(() => {
    if (inFlight.current || attempt.submission?.status !== 'infra_failed') return
    inFlight.current = true
    void perform({
      kind: 'retry',
      submissionID: attempt.submission.id,
      requestKey: requestKey(),
    })
  }, [attempt.submission, perform])

  const retryRequest = useCallback(() => {
    if (inFlight.current || state.status !== 'error' || !state.pending) return
    inFlight.current = true
    void perform(state.pending)
  }, [perform, state])

  return {
    error: state.status === 'error' ? state.message : null,
    hasRetryableRequest: state.status === 'error' && state.pending !== undefined,
    busy: state.status === 'sending',
    retryInfrastructure,
    retryRequest,
    submit,
  }
}

function requestKey(): string {
  return globalThis.crypto?.randomUUID?.() ?? `submission-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function errorText(error: unknown): string {
  if (error instanceof LearningApiError) return '本次提交暂时失败，请重试。'
  if (error instanceof Error) return error.message
  return '本次提交暂时失败，请重试。'
}
