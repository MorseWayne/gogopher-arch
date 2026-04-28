import { Terminal } from 'lucide-react';
import { SectionTitle } from '../common/SectionTitle';
import { formatDuration } from '../../lib/formatDuration';
import type { SandboxResponse } from '../../taskFeedback';
import styles from './FeedbackPanel.module.css';

interface ConsoleProps {
  output: SandboxResponse | null;
  error: string | null;
}

export function Console({ output, error }: ConsoleProps) {
  return (
    <section className={styles.consoleSection}>
      <div className={styles.consoleHeader}>
        <SectionTitle icon={Terminal} label="控制台" />
        <span className={styles.consoleDuration}>
          {output ? formatDuration(output.duration) : '--'}
        </span>
      </div>
      <div className={`${styles.consoleBody} ${error ? styles.hasError : ''}`}>
        {error && <pre className={styles.consoleError}>{error}</pre>}
        {output ? (
          <>
            {output.stdout && <pre>{output.stdout}</pre>}
            {output.stderr && <pre className={styles.consoleError}>{output.stderr}</pre>}
            <p className={styles.consoleMeta}>
              退出码：{output.exit_code} · 状态：{output.status.toUpperCase()}
            </p>
          </>
        ) : (
          <p className={styles.consolePlaceholder}>点击运行代码，查看沙盒输出。</p>
        )}
      </div>
    </section>
  );
}
