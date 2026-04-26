import { useMemo, useState } from 'react';
import Editor from '@monaco-editor/react';
import axios from 'axios';
import {
  AlertCircle,
  BookOpen,
  CheckCircle2,
  ClipboardCheck,
  Code2,
  GraduationCap,
  Play,
  Terminal,
} from 'lucide-react';
import './App.css';

interface SandboxResponse {
  stdout: string;
  stderr: string;
  status: string;
  duration: number;
  exit_code: number;
}

type FeedbackState = 'idle' | 'pass' | 'fail';

interface FeedbackItem {
  label: string;
  detail: string;
  state: FeedbackState;
}

const DEFAULT_CODE = `package main

import "fmt"

type User struct {
\tName  string
\tScore int
}

func buildScoreMap(users []User) map[string]int {
\tvar scores map[string]int
\tfor _, user := range users {
\t\tscores[user.Name] = user.Score
\t}
\treturn scores
}

func main() {
\tusers := []User{
\t\t{Name: "Ming", Score: 86},
\t\t{Name: "Yan", Score: 91},
\t}

\tscores := buildScoreMap(users)
\tfmt.Println("Ming 的分数:", scores["Ming"])
}
`;

const taskCriteria = [
  '程序可以成功运行，不再出现 nil map 写入 panic。',
  'buildScoreMap 返回包含所有用户分数的 map。',
  '不要修改 main 函数里的输入数据和输出语句。',
];

const lessonPoints = [
  'map 在写入前必须完成初始化。',
  'var scores map[string]int 声明的是 nil map，只能读，不能写。',
  'make(map[string]int, len(users)) 可以创建可写 map，并预留容量。',
];

const mentorHints = [
  '先定位 panic 行，再判断这个变量是否已经初始化。',
  '这类问题在实习任务里很常见：看起来类型对了，但零值不能直接写入。',
  '修复后再运行一次，确认 stdout 中出现 Ming 的分数。',
];

function getFeedback(output: SandboxResponse | null, error: string | null): FeedbackItem[] {
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
      {
        label: '任务检查',
        detail: '任务检查需要基于沙盒运行结果判断。',
        state: 'idle',
      },
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
      {
        label: '任务检查',
        detail: '修复 nil map 后，程序应输出 Ming 的分数。',
        state: 'idle',
      },
    ];
  }

  const succeeded = output.status === 'success' && output.exit_code === 0;
  const hasExpectedOutput = output.stdout.includes('Ming 的分数:');

  return [
    {
      label: '连接 Gateway',
      detail: '已收到沙盒执行结果。',
      state: 'pass',
    },
    {
      label: '运行结果',
      detail: succeeded ? '程序正常退出。' : '程序未正常退出，请查看 stderr。',
      state: succeeded ? 'pass' : 'fail',
    },
    {
      label: '任务检查',
      detail: hasExpectedOutput ? '已输出目标用户分数。' : '还没有看到预期输出。',
      state: hasExpectedOutput ? 'pass' : 'fail',
    },
  ];
}

function formatDuration(duration: number): string {
  if (duration <= 0) {
    return '--';
  }

  return `${(duration / 1_000_000).toFixed(2)}ms`;
}

function App() {
  const [code, setCode] = useState(DEFAULT_CODE);
  const [output, setOutput] = useState<SandboxResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const feedback = useMemo(() => getFeedback(output, error), [output, error]);

  const handleRun = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await axios.post<SandboxResponse>('http://localhost:8080/api/v1/execute', {
        id: `task-${Date.now()}`,
        code,
        language: 'go',
        timeout: 5,
      });
      setOutput(response.data);
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        const message =
          typeof err.response?.data === 'string'
            ? err.response.data
            : err.message || '无法连接到 Gateway 服务';
        setError(message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('无法连接到 Gateway 服务');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">
          <div className="brand-icon">
            <GraduationCap size={22} />
          </div>
          <div>
            <p className="eyebrow">Go 后端实习生 · 入职第一周</p>
            <h1>GoGopher Arch</h1>
          </div>
        </div>
        <button className="run-button" onClick={handleRun} disabled={loading}>
          <Play size={17} fill="currentColor" />
          {loading ? '运行中' : '运行代码'}
        </button>
      </header>

      <main className="workbench">
        <aside className="task-panel" aria-label="任务卡">
          <section className="panel-section hero-section">
            <div className="section-title">
              <ClipboardCheck size={16} />
              <span>任务卡</span>
            </div>
            <h2>Day 1：修复 nil map 写入</h2>
            <p>
              你的导师把一个用户分数统计函数交给你。当前代码会在运行时 panic，
              请定位原因并完成修复。
            </p>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <CheckCircle2 size={16} />
              <span>验收标准</span>
            </div>
            <ul className="check-list">
              {taskCriteria.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <BookOpen size={16} />
              <span>任务前小课</span>
            </div>
            <ul className="lesson-list">
              {lessonPoints.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </section>
        </aside>

        <section className="editor-panel" aria-label="代码编辑器">
          <div className="panel-toolbar">
            <div className="section-title">
              <Code2 size={16} />
              <span>main.go</span>
            </div>
            <span className="file-badge">Go 基础 Bug 修复</span>
          </div>
          <Editor
            height="100%"
            theme="vs-dark"
            defaultLanguage="go"
            value={code}
            onChange={(value) => setCode(value || '')}
            options={{
              fontSize: 14,
              minimap: { enabled: false },
              padding: { top: 18 },
              fontFamily: "'JetBrains Mono', 'Fira Code', ui-monospace, monospace",
              scrollBeyondLastLine: false,
            }}
          />
        </section>

        <aside className="feedback-panel" aria-label="任务反馈">
          <section className="panel-section">
            <div className="section-title">
              <ClipboardCheck size={16} />
              <span>任务反馈</span>
            </div>
            <div className="feedback-list">
              {feedback.map((item) => (
                <div className={`feedback-item ${item.state}`} key={item.label}>
                  <span className="feedback-dot" />
                  <div>
                    <strong>{item.label}</strong>
                    <p>{item.detail}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <AlertCircle size={16} />
              <span>导师提示</span>
            </div>
            <ul className="hint-list">
              {mentorHints.map((hint) => (
                <li key={hint}>{hint}</li>
              ))}
            </ul>
          </section>

          <section className="console-section">
            <div className="console-header">
              <div className="section-title">
                <Terminal size={16} />
                <span>控制台</span>
              </div>
              <span>{output ? formatDuration(output.duration) : '--'}</span>
            </div>
            <div className="console-body">
              {error && <pre className="console-error">{error}</pre>}
              {output ? (
                <>
                  {output.stdout && <pre>{output.stdout}</pre>}
                  {output.stderr && <pre className="console-error">{output.stderr}</pre>}
                  <p className="console-meta">
                    退出码：{output.exit_code} · 状态：{output.status.toUpperCase()}
                  </p>
                </>
              ) : (
                <p className="console-placeholder">点击运行代码，查看沙盒输出。</p>
              )}
            </div>
          </section>
        </aside>
      </main>
    </div>
  );
}

export default App;
