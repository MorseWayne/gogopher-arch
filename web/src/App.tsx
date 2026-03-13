import React, { useState } from 'react';
import Editor from "@monaco-editor/react";
import { Play, Activity, Terminal, Code2, AlertCircle } from 'lucide-react';
import axios from 'axios';

interface SandboxResponse {
  stdout: string;
  stderr: string;
  status: string;
  duration: number;
  exit_code: number;
}

const DEFAULT_CODE = `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("🚀 GoGopher Arch: 系统自检中...")
	
	// 这是一个典型的 Goroutine 泄露风险代码
	// 如果你不小心在循环中开启了没有结束条件的协程...
	for i := 0; i < 5; i++ {
		go func(id int) {
			fmt.Printf("Worker %d 启动并进入死循环...\n", id)
			for {
				time.Sleep(time.Second)
			}
		}(i)
	}

	time.Sleep(2 * time.Second)
	fmt.Println("✅ 任务执行完毕，观察右侧资源占用！")
}
`;

function App() {
  const [code, setCode] = useState(DEFAULT_CODE);
  const [output, setOutput] = useState<SandboxResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleRun = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.post('http://localhost:8080/api/v1/execute', {
        id: `task-${Date.now()}`,
        code: code,
        language: 'go',
        timeout: 5
      });
      setOutput(response.data);
    } catch (err: any) {
      setError(err.response?.data || err.message || "无法连接到 Gateway 服务");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col h-screen bg-[#1e1e1e] text-gray-300 font-sans">
      {/* 顶部导航 */}
      <header className="flex items-center justify-between px-6 py-3 border-bottom border-gray-700 bg-[#252526]">
        <div className="flex items-center gap-2">
          <div className="bg-blue-600 p-1.5 rounded">
            <Code2 size={20} className="text-white" />
          </div>
          <h1 className="text-lg font-bold text-white tracking-tight">GoGopher Arch <span className="text-xs font-normal text-gray-500 italic ml-1">v0.1-mvp</span></h1>
        </div>
        <div className="flex items-center gap-4">
          <button 
            onClick={handleRun}
            disabled={loading}
            className={`flex items-center gap-2 px-4 py-1.5 rounded text-sm font-medium transition-all ${loading ? 'bg-gray-700 cursor-not-allowed' : 'bg-green-600 hover:bg-green-700 text-white shadow-lg'}`}
          >
            <Play size={16} fill="currentColor" />
            {loading ? '运行中...' : '运行代码'}
          </button>
        </div>
      </header>

      {/* 主工作区 */}
      <main className="flex flex-1 overflow-hidden">
        {/* 编辑器区域 */}
        <div className="flex-1 flex flex-col border-r border-gray-800">
          <div className="bg-[#2d2d2d] px-4 py-2 text-xs uppercase tracking-wider text-gray-400 flex items-center gap-2">
            <Code2 size={14} /> main.go
          </div>
          <Editor
            height="100%"
            theme="vs-dark"
            defaultLanguage="go"
            value={code}
            onChange={(v) => setCode(v || "")}
            options={{
              fontSize: 14,
              minimap: { enabled: false },
              padding: { top: 20 },
              fontFamily: "'JetBrains Mono', 'Fira Code', monospace"
            }}
          />
        </div>

        {/* 控制台与反馈区域 */}
        <div className="w-[400px] flex flex-col bg-[#1e1e1e]">
          {/* 实时监控面板 (预览) */}
          <section className="p-4 border-b border-gray-800 bg-[#252526]">
            <div className="flex items-center gap-2 text-xs uppercase tracking-wider text-gray-400 mb-4">
              <Activity size={14} className="text-blue-500" /> 实时指标 (Metrics)
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="bg-[#1e1e1e] p-3 rounded border border-gray-800">
                <div className="text-[10px] text-gray-500 mb-1">活跃 Goroutines</div>
                <div className={`text-xl font-bold ${output?.status === 'timeout' ? 'text-red-500' : 'text-blue-400'}`}>
                  {output ? (output.status === 'success' ? '6' : '∞') : '--'}
                </div>
              </div>
              <div className="bg-[#1e1e1e] p-3 rounded border border-gray-800">
                <div className="text-[10px] text-gray-500 mb-1">执行耗时</div>
                <div className="text-xl font-bold text-yellow-400">
                  {output ? `${(output.duration / 1000000).toFixed(2)}ms` : '--'}
                </div>
              </div>
            </div>
          </section>

          {/* 控制台终端 */}
          <section className="flex-1 flex flex-col min-h-0">
            <div className="bg-[#2d2d2d] px-4 py-2 text-xs uppercase tracking-wider text-gray-400 flex items-center gap-2">
              <Terminal size={14} /> 控制台 (Console)
            </div>
            <div className="flex-1 p-4 font-mono text-sm overflow-y-auto bg-black text-green-400">
              {error && (
                <div className="flex items-start gap-2 text-red-400 mb-2">
                  <AlertCircle size={16} className="shrink-0 mt-0.5" />
                  <span>{error}</span>
                </div>
              )}
              {output ? (
                <>
                  <pre className="whitespace-pre-wrap mb-2">{output.stdout}</pre>
                  {output.stderr && <pre className="text-red-400 whitespace-pre-wrap">{output.stderr}</pre>}
                  <div className="mt-4 pt-2 border-t border-gray-800 text-gray-500 text-xs">
                    程序退出码: {output.exit_code} | 状态: {output.status.toUpperCase()}
                  </div>
                </>
              ) : (
                <div className="text-gray-600 italic">点击“运行代码”查看输出结果...</div>
              )}
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}

export default App;
