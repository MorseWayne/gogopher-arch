import { AlertCircle, BookOpen, ClipboardCheck } from 'lucide-react';
import { SectionTitle } from '../common/SectionTitle';
import { FeedbackList } from './FeedbackList';
import { Console } from './Console';
import type { FeedbackPanelProps } from '../../types/workbench';
import styles from './FeedbackPanel.module.css';

export function FeedbackPanel({
  feedback,
  currentTaskPassed,
  mentorHints,
  review,
  output,
  error,
}: FeedbackPanelProps) {
  return (
    <aside className={styles.panel} aria-label="任务反馈">
      <section className={styles.section}>
        <SectionTitle icon={ClipboardCheck} label="任务反馈" />
        <div className={styles.feedbackSummary}>
          {currentTaskPassed ? '本任务已通过。' : '运行代码后查看任务检查。'}
        </div>
        <FeedbackList items={feedback} />
      </section>

      <section className={styles.section}>
        <SectionTitle icon={AlertCircle} label="导师提示" />
        <ul className={styles.hintList}>
          {mentorHints.map((hint) => (
            <li key={hint}>{hint}</li>
          ))}
        </ul>
      </section>

      <section className={styles.section}>
        <SectionTitle icon={BookOpen} label="任务后复盘" />
        <ul className={styles.reviewList}>
          {review.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </section>

      <Console output={output} error={error} />
    </aside>
  );
}
