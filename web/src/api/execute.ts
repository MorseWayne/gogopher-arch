import { apiClient } from './client';
import type { SandboxRequest, SandboxResponse } from './types';

export async function executeCode(req: SandboxRequest): Promise<SandboxResponse> {
  const { data } = await apiClient.post<SandboxResponse>('/execute', req);
  return data;
}