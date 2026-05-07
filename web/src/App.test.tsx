import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it, vi } from 'vitest';
import App from './App';

vi.mock('@monaco-editor/react', () => ({
  default: () => <div data-testid="editor" />,
}));

describe('App', () => {
  it('renders the first-week task line workbench', () => {
    const markup = renderToStaticMarkup(<App />);

    expect(markup).toContain('Day 0');
    expect(markup).toContain('第一次运行 Go 代码');
    expect(markup).toContain('当前任务');
    expect(markup).toContain('任务反馈');
    expect(markup).toContain('任务后复盘');
  });
});
