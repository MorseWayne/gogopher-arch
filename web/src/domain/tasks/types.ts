// Type definitions for task domain layer

export interface SandboxResponse {
  stdout: string;
  stderr: string;
  status: string;
  duration: number;
  exit_code: number;
}

export type FeedbackState = 'idle' | 'pass' | 'fail';

export interface FeedbackItem {
  label: string;
  detail: string;
  state: FeedbackState;
}

export type TaskCheck =
  | {
      type: 'exitSuccess';
      label: string;
      passDetail: string;
      failDetail: string;
    }
  | {
      type: 'stdoutIncludes';
      label: string;
      passDetail: string;
      failDetail: string;
      value: string;
    }
  | {
      type: 'stdoutRegex';
      label: string;
      passDetail: string;
      failDetail: string;
      pattern: string;
      flags?: string;
    }
  | {
      type: 'stderrExcludes';
      label: string;
      passDetail: string;
      failDetail: string;
      value: string;
    };

export interface InternTask {
  id: string;
  day: number;
  title: string;
  track: string;
  summary: string;
  background: string;
  objective: string;
  starterCode: string;
  criteria: string[];
  lesson: string[];
  mentorHints: string[];
  review: string[];
  checks: TaskCheck[];
}