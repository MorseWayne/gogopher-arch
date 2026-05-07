import { TaskContent } from './TaskContent';
import type { TaskPanelProps } from '@/domain/workbench';
import styles from './TaskPanel.module.css';

export function TaskPanel({ task }: TaskPanelProps) {
  return (
    <aside className={styles.panel} aria-label="任务详情">
      <TaskContent task={task} />
    </aside>
  );
}
