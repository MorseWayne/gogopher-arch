import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it, vi } from 'vitest';
import App from './App';

vi.mock('@monaco-editor/react', () => ({
  default: () => <div data-testid="editor" />,
}));

describe('App', () => {
  it('renders the first-week task line workbench', () => {
    const markup = renderToStaticMarkup(<App />);

    expect(markup).toContain('任务列表');
    expect(markup).toContain('Day 0');
    expect(markup).toContain('第一次运行 Go 代码');
    expect(markup).toContain('任务后复盘');
    expect(markup).toContain('GoGopher Day 0 sandbox ready');
  });
});
