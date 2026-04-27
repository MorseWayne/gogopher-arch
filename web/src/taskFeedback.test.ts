import { describe, expect, it } from 'vitest';
import {
  didPassTask,
  evaluateTaskChecks,
  type SandboxResponse,
  type TaskCheck,
} from './taskFeedback';

const checks: TaskCheck[] = [
  {
    type: 'stdoutIncludes',
    label: 'stdout phrase',
    passDetail: 'stdout contains the expected phrase.',
    failDetail: 'stdout does not contain the expected phrase.',
    value: 'hello intern',
  },
  {
    type: 'stderrExcludes',
    label: 'no panic',
    passDetail: 'stderr does not include panic.',
    failDetail: 'stderr still includes panic.',
    value: 'panic:',
  },
];

function sandbox(overrides: Partial<SandboxResponse> = {}): SandboxResponse {
  return {
    stdout: 'hello intern\n',
    stderr: '',
    status: 'success',
    duration: 1200000,
    exit_code: 0,
    ...overrides,
  };
}

describe('evaluateTaskChecks', () => {
  it('returns idle feedback before the first run', () => {
    const feedback = evaluateTaskChecks(null, null, checks);

    expect(feedback.map((item) => item.state)).toEqual(['idle', 'idle', 'idle', 'idle']);
    expect(feedback[0].label).toBe('连接 Gateway');
    expect(feedback[2].label).toBe('stdout phrase');
  });

  it('returns connection failure feedback when the gateway request fails', () => {
    const feedback = evaluateTaskChecks(null, 'network error', checks);

    expect(feedback[0]).toMatchObject({
      label: '连接 Gateway',
      state: 'fail',
    });
    expect(feedback[2].state).toBe('idle');
  });

  it('passes task checks when the sandbox succeeds and all checks match', () => {
    const feedback = evaluateTaskChecks(sandbox(), null, checks);

    expect(feedback.map((item) => item.state)).toEqual(['pass', 'pass', 'pass', 'pass']);
    expect(didPassTask(sandbox(), null, checks)).toBe(true);
  });

  it('fails task checks when stdout or stderr does not match', () => {
    const output = sandbox({
      stdout: 'wrong output\n',
      stderr: 'panic: assignment to entry in nil map',
      status: 'error',
      exit_code: 1,
    });

    const feedback = evaluateTaskChecks(output, null, checks);

    expect(feedback.map((item) => item.state)).toEqual(['pass', 'fail', 'fail', 'fail']);
    expect(didPassTask(output, null, checks)).toBe(false);
  });
});
