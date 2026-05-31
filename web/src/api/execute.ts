import type { SandboxRequest, SandboxResponse } from './types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

export async function executeCode(req: SandboxRequest): Promise<SandboxResponse> {
  const response = await fetch(`${API_BASE_URL}/execute`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(req),
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Gateway request failed with ${response.status}`);
  }

  return response.json();
}
