import Editor from '@monaco-editor/react';
import { EditorToolbar } from './EditorToolbar';
import type { EditorPanelProps } from '../../types/workbench';
import styles from './EditorPanel.module.css';

export function EditorPanel({ code, onChange, track }: EditorPanelProps) {
  return (
    <section className={styles.panel} aria-label="代码编辑器">
      <EditorToolbar track={track} />
      <div className={styles.editorWrapper}>
        <Editor
          height="100%"
          theme="vs-dark"
          defaultLanguage="go"
          value={code}
          onChange={(value) => onChange(value || '')}
          options={{
            fontSize: 14,
            fontFamily: "'JetBrains Mono', 'Fira Code', ui-monospace, monospace",
            minimap: { enabled: true, scale: 1 },
            padding: { top: 18 },
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            insertSpaces: true,
          }}
        />
      </div>
    </section>
  );
}
