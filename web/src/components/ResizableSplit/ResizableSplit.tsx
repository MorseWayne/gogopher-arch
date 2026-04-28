import { useState } from 'react';
import { useResizable } from '../../hooks/useResizable';
import type { ResizableSplitProps } from '../../types/workbench';
import styles from './ResizableSplit.module.css';

export function ResizableSplit({
  left,
  center,
  right,
}: ResizableSplitProps) {
  const [leftCollapsed, setLeftCollapsed] = useState(false);
  const [rightCollapsed, setRightCollapsed] = useState(false);

  const leftPanel = useResizable({
    initialWidth: 300,
    minWidth: 200,
    maxWidth: 500,
    storageKey: 'gogopher-panel-left',
  });

  const rightPanel = useResizable({
    initialWidth: 340,
    minWidth: 240,
    maxWidth: 500,
    storageKey: 'gogopher-panel-right',
  });

  const leftWidth = leftCollapsed ? 32 : leftPanel.width;
  const rightWidth = rightCollapsed ? 32 : rightPanel.width;

  return (
    <div
      className={styles.split}
      style={{
        gridTemplateColumns: `${leftWidth}px 4px 1fr 4px ${rightWidth}px`,
      }}
    >
      <div className={`${styles.panel} ${leftCollapsed ? styles.collapsed : ''}`}>
        {!leftCollapsed && left}
        {leftCollapsed && (
          <button
            className={styles.expandBtn}
            onClick={() => setLeftCollapsed(false)}
            aria-label="展开任务面板"
          >
            ›
          </button>
        )}
      </div>

      <div
        className={styles.handle}
        onMouseDown={leftCollapsed ? undefined : leftPanel.startResize('left')}
        title={leftCollapsed ? undefined : '拖拽调整宽度'}
      />

      <div className={styles.center}>{center}</div>

      <div
        className={styles.handle}
        onMouseDown={rightCollapsed ? undefined : rightPanel.startResize('right')}
        title={rightCollapsed ? undefined : '拖拽调整宽度'}
      />

      <div className={`${styles.panel} ${rightCollapsed ? styles.collapsed : ''}`}>
        {!rightCollapsed && right}
        {rightCollapsed && (
          <button
            className={styles.expandBtn}
            onClick={() => setRightCollapsed(false)}
            aria-label="展开反馈面板"
          >
            ‹
          </button>
        )}
      </div>
    </div>
  );
}
