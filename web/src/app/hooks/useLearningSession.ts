import { useCallback, useEffect, useState } from 'react'

import { bootstrapLearningSession } from '../../api/learning'
import type { SessionResponse } from '../../api/learning'

type SessionState =
  | { status: 'loading' }
  | { status: 'ready'; session: SessionResponse }
  | { status: 'error'; error: unknown }

export type LearningSessionState = SessionState & {
  retry: () => void
}

export function useLearningSession(): LearningSessionState {
  const [generation, setGeneration] = useState(0)
  const [state, setState] = useState<SessionState>({ status: 'loading' })

  useEffect(() => {
    let current = true
    setState({ status: 'loading' })
    void bootstrapLearningSession().then(
      (session) => {
        if (current) setState({ status: 'ready', session })
      },
      (error: unknown) => {
        if (current) setState({ status: 'error', error })
      },
    )
    return () => {
      current = false
    }
  }, [generation])

  const retry = useCallback(() => setGeneration((value) => value + 1), [])
  return { ...state, retry }
}
