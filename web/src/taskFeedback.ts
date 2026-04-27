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

function didSandboxSucceed(output: SandboxResponse): boolean {
  return output.status === 'success' && output.exit_code === 0;
}

function idleCheckFeedback(check: TaskCheck): FeedbackItem {
  return {
    label: check.label,
    detail: '运行当前任务后查看检查结果。',
    state: 'idle',
  };
}

function evaluateSingleCheck(check: TaskCheck, output: SandboxResponse): FeedbackItem {
  let passed = false;

  if (check.type === 'exitSuccess') {
    passed = didSandboxSucceed(output);
  }

  if (check.type === 'stdoutIncludes') {
    passed = output.stdout.includes(check.value);
  }

  if (check.type === 'stdoutRegex') {
    passed = new RegExp(check.pattern, check.flags).test(output.stdout);
  }

  if (check.type === 'stderrExcludes') {
    passed = !output.stderr.includes(check.value);
  }

  return {
    label: check.label,
    detail: passed ? check.passDetail : check.failDetail,
    state: passed ? 'pass' : 'fail',
  };
}

export function evaluateTaskChecks(
  output: SandboxResponse | null,
  error: string | null,
  checks: TaskCheck[],
): FeedbackItem[] {
  if (error) {
    return [
      {
        label: '连接 Gateway',
        detail: '前端无法连接到本地 Gateway，请确认后端服务已启动。',
        state: 'fail',
      },
      {
        label: '运行结果',
        detail: '等待 Gateway 恢复后重新运行。',
        state: 'idle',
      },
      ...checks.map(idleCheckFeedback),
    ];
  }

  if (!output) {
    return [
      {
        label: '连接 Gateway',
        detail: '等待第一次运行。',
        state: 'idle',
      },
      {
        label: '运行结果',
        detail: '点击运行代码后查看 stdout 和 stderr。',
        state: 'idle',
      },
      ...checks.map(idleCheckFeedback),
    ];
  }

  const runSucceeded = didSandboxSucceed(output);

  return [
    {
      label: '连接 Gateway',
      detail: '已收到沙盒执行结果。',
      state: 'pass',
    },
    {
      label: '运行结果',
      detail: runSucceeded ? '程序正常退出。' : '程序未正常退出，请查看 stderr。',
      state: runSucceeded ? 'pass' : 'fail',
    },
    ...checks.map((check) => evaluateSingleCheck(check, output)),
  ];
}

export function didPassTask(
  output: SandboxResponse | null,
  error: string | null,
  checks: TaskCheck[],
): boolean {
  if (error || !output || !didSandboxSucceed(output)) {
    return false;
  }

  return checks.every((check) => evaluateSingleCheck(check, output).state === 'pass');
}
