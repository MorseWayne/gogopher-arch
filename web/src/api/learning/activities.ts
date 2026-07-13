import { learningRequest } from './client'
import type { ActivityResponse, CapabilityResponse, NextResponse } from './types'

export function getActivity(activityID: string, version = 1): Promise<ActivityResponse> {
  return learningRequest(`/activities/${encodeURIComponent(activityID)}?version=${version}`)
}

export function getCapability(capabilityID: string): Promise<CapabilityResponse> {
  return learningRequest(`/capabilities/${encodeURIComponent(capabilityID)}`)
}

export function getNextRecommendation(): Promise<NextResponse> {
  return learningRequest('/next')
}
