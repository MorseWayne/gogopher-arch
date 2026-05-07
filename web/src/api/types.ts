export interface SandboxRequest {
  id: string;
  code: string;
  language: string;
  timeout: number;
}

export interface SandboxResponse {
  id: string;
  stdout: string;
  stderr: string;
  exit_code: number;
  duration: number;
  status: string;
}