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
  RotateCcw,
  Terminal,
} from 'lucide-react';
import './App.css';
import { defaultTaskId, findTaskById, internshipTasks } from './tasks';
import {
  didPassTask,
  evaluateTaskChecks,
  type SandboxResponse,
} from './taskFeedback';

function formatDuration(duration: number): string {
  if (duration <= 0) {
    return '--';
  }

  return `${(duration / 1_000_000).toFixed(2)}ms`;
}

function App() {
  const [selectedTaskId, setSelectedTaskId] = useState(defaultTaskId);
  const selectedTask = findTaskById(selectedTaskId);
  const [code, setCode] = useState(selectedTask.starterCode);
  const [output, setOutput] = useState<SandboxResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [taskResults, setTaskResults] = useState<Record<string, 'pass' | 'fail'>>({});

  const feedback = useMemo(
    () => evaluateTaskChecks(output, error, selectedTask.checks),
    [output, error, selectedTask],
  );

  const currentTaskPassed = didPassTask(output, error, selectedTask.checks);

  const handleSelectTask = (taskId: string) => {
    const nextTask = findTaskById(taskId);
    setSelectedTaskId(nextTask.id);
    setCode(nextTask.starterCode);
    setOutput(null);
    setError(null);
  };

  const handleResetCode = () => {
    setCode(selectedTask.starterCode);
    setOutput(null);
    setError(null);
  };

  const handleRun = async () => {
    setLoading(true);
    setError(null);
    setOutput(null);

    try {
      const response = await axios.post<SandboxResponse>('http://localhost:8080/api/v1/execute', {
        id: `${selectedTask.id}-${Date.now()}`,
        code,
        language: 'go',
        timeout: 5,
      });
      const nextOutput = response.data;
      setOutput(nextOutput);
      setTaskResults((results) => ({
        ...results,
        [selectedTask.id]: didPassTask(nextOutput, null, selectedTask.checks) ? 'pass' : 'fail',
      }));
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
      setTaskResults((results) => ({
        ...results,
        [selectedTask.id]: 'fail',
      }));
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
        <div className="topbar-actions">
          <button className="ghost-button" onClick={handleResetCode} disabled={loading}>
            <RotateCcw size={16} />
            重置代码
          </button>
          <button className="run-button" onClick={handleRun} disabled={loading}>
            <Play size={17} fill="currentColor" />
            {loading ? '运行中' : '运行代码'}
          </button>
        </div>
      </header>

      <main className="workbench">
        <aside className="task-panel" aria-label="任务卡">
          <section className="panel-section task-nav-section">
            <div className="section-title">
              <ClipboardCheck size={16} />
              <span>任务列表</span>
            </div>
            <div className="task-list">
              {internshipTasks.map((task) => {
                const result = taskResults[task.id];
                const isSelected = task.id === selectedTask.id;

                return (
                  <button
                    className={`task-list-item ${isSelected ? 'selected' : ''} ${result || ''}`}
                    key={task.id}
                    onClick={() => handleSelectTask(task.id)}
                    type="button"
                  >
                    <span className="task-day">Day {task.day}</span>
                    <span className="task-list-title">{task.title}</span>
                    <span className="task-list-track">{task.track}</span>
                  </button>
                );
              })}
            </div>
          </section>

          <section className="panel-section hero-section">
            <div className="section-title">
              <ClipboardCheck size={16} />
              <span>任务卡</span>
            </div>
            <h2>
              Day {selectedTask.day}：{selectedTask.title}
            </h2>
            <p>{selectedTask.background}</p>
            <p className="objective">{selectedTask.objective}</p>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <CheckCircle2 size={16} />
              <span>验收标准</span>
            </div>
            <ul className="check-list">
              {selectedTask.criteria.map((item) => (
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
              {selectedTask.lesson.map((item) => (
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
            <span className="file-badge">{selectedTask.track}</span>
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
            <div className="feedback-summary">
              {currentTaskPassed ? '本任务已通过。' : '运行代码后查看任务检查。'}
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
              {selectedTask.mentorHints.map((hint) => (
                <li key={hint}>{hint}</li>
              ))}
            </ul>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <BookOpen size={16} />
              <span>任务后复盘</span>
            </div>
            <ul className="review-list">
              {selectedTask.review.map((item) => (
                <li key={item}>{item}</li>
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
