import { learningRequest } from './client'
import type { SessionResponse } from './types'

export function bootstrapLearningSession(): Promise<SessionResponse> {
  return learningRequest('/session', { method: 'POST' })
}
