import { GraduationCap, Play, RotateCcw } from 'lucide-react';
import type { TopBarProps } from '@/domain/workbench';
import styles from './TopBar.module.css';

export function TopBar({ onReset, onRun, loading }: TopBarProps) {
  return (
    <header className={styles.topbar}>
      <div className={styles.brand}>
        <div className={styles.brandIcon}>
          <GraduationCap size={22} />
        </div>
        <div>
          <p className={styles.eyebrow}>Go 后端实习生 · 入职第一周</p>
          <h1 className={styles.brandTitle}>GoGopher Arch</h1>
        </div>
      </div>
      <div className={styles.actions}>
        <button className={styles.ghostButton} onClick={onReset} disabled={loading}>
          <RotateCcw size={16} />
          重置代码
        </button>
        <button className={styles.runButton} onClick={onRun} disabled={loading}>
          {loading ? (
            <span className={styles.spinner} />
          ) : (
            <Play size={17} fill="currentColor" />
          )}
          {loading ? '运行中' : '运行代码'}
        </button>
      </div>
    </header>
  );
}
