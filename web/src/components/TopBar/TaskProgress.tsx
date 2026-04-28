import { Check } from 'lucide-react';
import type { TaskProgressProps } from '../../types/workbench';
import styles from './TaskProgress.module.css';

export function TaskProgress({ tasks, selectedTaskId, taskResults, onSelectTask }: TaskProgressProps) {
  return (
    <div className={styles.progressBar}>
      {tasks.map((task, index) => {
        const result = taskResults[task.id];
        const isSelected = task.id === selectedTaskId;
        const isPassed = result === 'pass';

        return (
          <button
            key={task.id}
            className={`${styles.dayItem} ${isSelected ? styles.selected : ''} ${isPassed ? styles.passed : ''}`}
            onClick={() => onSelectTask(task.id)}
            type="button"
          >
            <span className={styles.dayLabel}>Day {task.day}</span>
            {isPassed && <Check size={12} className={styles.checkIcon} />}
            {index < tasks.length - 1 && <span className={styles.arrow}>→</span>}
          </button>
        );
      })}
    </div>
  );
}
