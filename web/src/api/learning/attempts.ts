import { jsonBody, learningRequest } from './client'
import type {
  AssistanceEventResponse,
  AssistanceEventType,
  AttemptResponse,
  CreateAttemptRequest,
  ExecuteAttemptRequest,
  Execution,
  HintResponse,
  SaveWorkspaceRequest,
  SubmissionResponse,
  SubmitAttemptRequest,
} from './types'

export function createAttempt(request: CreateAttemptRequest): Promise<AttemptResponse> {
  return learningRequest('/attempts', { method: 'POST', ...jsonBody(request) })
}

export function claimReviewAttempt(reviewItemID: string): Promise<AttemptResponse> {
  return learningRequest(`/review-items/${encodeURIComponent(reviewItemID)}/attempts`, { method: 'POST' })
}

export function getAttempt(attemptID: string): Promise<AttemptResponse> {
  return learningRequest(`/attempts/${encodeURIComponent(attemptID)}`)
}

export function saveWorkspace(attemptID: string, request: SaveWorkspaceRequest): Promise<AttemptResponse> {
  return learningRequest(`/attempts/${encodeURIComponent(attemptID)}/workspace`, {
    method: 'PUT',
    ...jsonBody(request),
  })
}

export function executeAttempt(attemptID: string, request: ExecuteAttemptRequest): Promise<Execution> {
  return learningRequest(`/attempts/${encodeURIComponent(attemptID)}/execute`, {
    method: 'POST',
    ...jsonBody(request),
  })
}

export function submitAttempt(attemptID: string, request: SubmitAttemptRequest): Promise<SubmissionResponse> {
  return learningRequest(`/attempts/${encodeURIComponent(attemptID)}/submit`, {
    method: 'POST',
    ...jsonBody(request),
  })
}

export function retrySubmission(submissionID: string, requestKey: string): Promise<SubmissionResponse> {
  return learningRequest(`/submissions/${encodeURIComponent(submissionID)}/retry`, {
    method: 'POST',
    ...jsonBody({ request_key: requestKey }),
  })
}

export function recordAssistance(
  attemptID: string,
  eventKey: string,
  eventType: AssistanceEventType,
  payload: Record<string, unknown> = {},
): Promise<AssistanceEventResponse> {
  return learningRequest(`/attempts/${encodeURIComponent(attemptID)}/assistance-events`, {
    method: 'POST',
    ...jsonBody({ event_key: eventKey, event_type: eventType, payload }),
  })
}

export function revealHint(attemptID: string, hintID: string, eventKey: string): Promise<HintResponse> {
  return learningRequest(`/attempts/${encodeURIComponent(attemptID)}/hints/${encodeURIComponent(hintID)}/reveal`, {
    method: 'POST',
    ...jsonBody({ event_key: eventKey }),
  })
}
