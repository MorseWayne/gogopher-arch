import { Code2 } from 'lucide-react';
import { SectionTitle } from '../common/SectionTitle';
import styles from './EditorPanel.module.css';

interface EditorToolbarProps {
  track: string;
}

export function EditorToolbar({ track }: EditorToolbarProps) {
  return (
    <div className={styles.toolbar}>
      <div className={styles.toolbarLeft}>
        <div className={styles.fileTab}>
          <SectionTitle icon={Code2} label="main.go" />
        </div>
      </div>
      <div className={styles.toolbarRight}>
        <span className={styles.badge}>{track}</span>
        <span className={styles.meta}>UTF-8 · Go</span>
      </div>
    </div>
  );
}
