import type { LearningErrorResponse } from './types'

const LEARNING_API_ROOT = '/api/v1/learning'

function isLearningErrorResponse(value: unknown): value is LearningErrorResponse {
  if (typeof value !== 'object' || value === null || !('error' in value)) {
    return false
  }
  const error = value.error
  return typeof error === 'object' && error !== null && 'code' in error && 'message' in error
    && typeof error.code === 'string' && typeof error.message === 'string'
}

export class LearningApiError extends Error {
  readonly status: number
  readonly code: string
  readonly payload: LearningErrorResponse

  constructor(status: number, payload: LearningErrorResponse) {
    super(payload.error.message)
    this.name = 'LearningApiError'
    this.status = status
    this.code = payload.error.code
    this.payload = payload
  }
}

export class LearningProtocolError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'LearningProtocolError'
    this.status = status
  }
}

export async function learningRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Accept', 'application/json')
  if (init.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(`${LEARNING_API_ROOT}${path}`, {
    ...init,
    credentials: 'same-origin',
    headers,
  })

  const body = await readJSON(response)
  if (!response.ok) {
    if (isLearningErrorResponse(body)) {
      throw new LearningApiError(response.status, body)
    }
    throw new LearningProtocolError(response.status, `Learning API returned HTTP ${response.status} without a domain error`)
  }
  return body as T
}

async function readJSON(response: Response): Promise<unknown> {
  try {
    return await response.json()
  } catch {
    throw new LearningProtocolError(response.status, 'Learning API returned invalid JSON')
  }
}

export function jsonBody(value: unknown): Pick<RequestInit, 'body'> {
  return { body: JSON.stringify(value) }
}
