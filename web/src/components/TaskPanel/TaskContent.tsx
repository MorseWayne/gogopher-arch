import { BookOpen, CheckCircle2, ClipboardCheck } from 'lucide-react';
import { SectionTitle } from '../common/SectionTitle';
import type { TaskPanelProps } from '@/domain/workbench';
import styles from './TaskPanel.module.css';

export function TaskContent({ task }: TaskPanelProps) {
  return (
    <>
      <section className={`${styles.section} ${styles.heroSection}`}>
        <SectionTitle icon={ClipboardCheck} label="当前任务" />
        <h2 className={styles.heading}>
          Day {task.day}：{task.title}
        </h2>
        <p className={styles.background}>{task.background}</p>
        <p className={styles.objective}>{task.objective}</p>
      </section>

      <section className={styles.section}>
        <SectionTitle icon={CheckCircle2} label="验收标准" />
        <ul className={styles.checkList}>
          {task.criteria.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </section>

      <section className={styles.section}>
        <SectionTitle icon={BookOpen} label="任务前小课" />
        <ul className={styles.lessonList}>
          {task.lesson.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </section>
    </>
  );
}
