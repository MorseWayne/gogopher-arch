import { useMemo, useState } from 'react';
import { TopBar } from './components/TopBar/TopBar';
import { TaskProgress } from './components/TopBar/TaskProgress';
import { TaskPanel } from './components/TaskPanel/TaskPanel';
import { EditorPanel } from './components/EditorPanel/EditorPanel';
import { FeedbackPanel } from './components/FeedbackPanel/FeedbackPanel';
import { ResizableSplit } from './components/ResizableSplit/ResizableSplit';
import { useMediaQuery } from './hooks/useMediaQuery';
import { defaultTaskId, findTaskById, internshipTasks } from './tasks';
import { didPassTask, evaluateTaskChecks, type SandboxResponse } from './taskFeedback';
import './index.css';
import styles from './App.module.css';
import axios from 'axios';

function App() {
  const [selectedTaskId, setSelectedTaskId] = useState(defaultTaskId);
  const selectedTask = findTaskById(selectedTaskId);
  const [code, setCode] = useState(selectedTask.starterCode);
  const [output, setOutput] = useState<SandboxResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [taskResults, setTaskResults] = useState<Record<string, 'pass' | 'fail'>>({});
  const [mobileTab, setMobileTab] = useState<'task' | 'editor' | 'feedback'>('editor');

  const isMobile = useMediaQuery('(max-width: 959px)');

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
      const response = await axios.post<SandboxResponse>('/api/v1/execute', {
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

  const taskPanelNode = <TaskPanel task={selectedTask} />;
  const editorPanelNode = (
    <EditorPanel code={code} onChange={setCode} track={selectedTask.track} />
  );
  const feedbackPanelNode = (
    <FeedbackPanel
      feedback={feedback}
      currentTaskPassed={currentTaskPassed}
      mentorHints={selectedTask.mentorHints}
      review={selectedTask.review}
      output={output}
      error={error}
    />
  );

  return (
    <div className={styles.appShell}>
      <TopBar onReset={handleResetCode} onRun={handleRun} loading={loading} />
      <TaskProgress
        tasks={internshipTasks}
        selectedTaskId={selectedTask.id}
        taskResults={taskResults}
        onSelectTask={handleSelectTask}
      />

      {isMobile ? (
        <main className={styles.mobileMain}>
          {mobileTab === 'task' && taskPanelNode}
          {mobileTab === 'editor' && editorPanelNode}
          {mobileTab === 'feedback' && feedbackPanelNode}

          <nav className={styles.mobileTabBar}>
            <button
              className={mobileTab === 'task' ? styles.activeTab : ''}
              onClick={() => setMobileTab('task')}
            >
              任务
            </button>
            <button
              className={mobileTab === 'editor' ? styles.activeTab : ''}
              onClick={() => setMobileTab('editor')}
            >
              编辑
            </button>
            <button
              className={mobileTab === 'feedback' ? styles.activeTab : ''}
              onClick={() => setMobileTab('feedback')}
            >
              反馈
            </button>
          </nav>
        </main>
      ) : (
        <main className={styles.desktopMain}>
          <ResizableSplit
            left={taskPanelNode}
            center={editorPanelNode}
            right={feedbackPanelNode}
          />
        </main>
      )}
    </div>
  );
}

export default App;
