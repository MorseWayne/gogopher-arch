import { learningRequest } from './client'
import type { ActivityResponse, CapabilityResponse, NextResponse } from './types'

export function getActivity(activityID: string, version = 1, releaseID = ''): Promise<ActivityResponse> {
  const query = new URLSearchParams({ version: String(version) })
  if (releaseID) query.set('release_id', releaseID)
  return learningRequest(`/activities/${encodeURIComponent(activityID)}?${query}`)
}

export function getCapability(capabilityID: string, version?: number, releaseID = ''): Promise<CapabilityResponse> {
  const query = new URLSearchParams()
  if (version !== undefined) query.set('version', String(version))
  if (releaseID) query.set('release_id', releaseID)
  const encoded = query.toString()
  const suffix = encoded ? `?${encoded}` : ''
  return learningRequest(`/capabilities/${encodeURIComponent(capabilityID)}${suffix}`)
}

export function getNextRecommendation(): Promise<NextResponse> {
  return learningRequest('/next')
}
