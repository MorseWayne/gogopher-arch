import type { FeedbackItem } from '@/domain/tasks';
import styles from './FeedbackPanel.module.css';

interface FeedbackListProps {
  items: FeedbackItem[];
}

export function FeedbackList({ items }: FeedbackListProps) {
  return (
    <div className={styles.feedbackList}>
      {items.map((item) => (
        <div className={`${styles.feedbackItem} ${styles[item.state]}`} key={item.label}>
          <span className={styles.feedbackDot} />
          <div>
            <strong>{item.label}</strong>
            <p>{item.detail}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
