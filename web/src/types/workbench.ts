import type { InternTask } from '../tasks';
import type { FeedbackItem, SandboxResponse } from '../taskFeedback';

export interface TopBarProps {
  onReset: () => void;
  onRun: () => void;
  loading: boolean;
}

export interface TaskProgressProps {
  tasks: InternTask[];
  selectedTaskId: string;
  taskResults: Record<string, 'pass' | 'fail'>;
  onSelectTask: (taskId: string) => void;
}

export interface TaskPanelProps {
  task: InternTask;
}

export interface EditorPanelProps {
  code: string;
  onChange: (value: string) => void;
  track: string;
}

export interface FeedbackPanelProps {
  feedback: FeedbackItem[];
  currentTaskPassed: boolean;
  mentorHints: string[];
  review: string[];
  output: SandboxResponse | null;
  error: string | null;
}

export interface ResizableSplitProps {
  left: React.ReactNode;
  center: React.ReactNode;
  right: React.ReactNode;
}
