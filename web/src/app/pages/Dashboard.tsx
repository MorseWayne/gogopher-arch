import { Link, useSearchParams } from "react-router";
import { useState, useEffect, useRef } from "react";
import { Play, Code2, Terminal, Cpu, Save, Settings, ChevronRight, FileCode2, FlaskConical, LayoutDashboard, Bot, FolderOpen, FileText, Send, Activity, AlertTriangle } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { executeCode } from "../../api/execute";
import type { SandboxResponse } from "../../api/types";
import { getMissionBySlug, missions, type Mission } from "../data/missions";

export function Dashboard() {
  const [searchParams] = useSearchParams();
  const mission = getMissionBySlug(searchParams.get("mission"));
  const [activeTab, setActiveTab] = useState<"code" | "console" | "ai">("code");
  const [isPlaying, setIsPlaying] = useState(false);
  const [cpuUsage, setCpuUsage] = useState(12);
  const [memUsage, setMemUsage] = useState(45);
  const [chatInput, setChatInput] = useState("");
  const [runResult, setRunResult] = useState<SandboxResponse | null>(null);
  const [runError, setRunError] = useState("");
  const endOfMessagesRef = useRef<HTMLDivElement>(null);
  const lineNumbers = mission.starterCode.split("\n").map((_, i) => i + 1);

  useEffect(() => {
    setRunResult(null);
    setRunError("");
    setActiveTab("code");
  }, [mission.slug]);

  useEffect(() => {
    const interval = setInterval(() => {
      setCpuUsage(prev => isPlaying ? Math.min(prev + Math.random() * 20, 99) : Math.max(prev - Math.random() * 10, 5));
      setMemUsage(prev => isPlaying ? Math.min(prev + Math.random() * 15, 99) : Math.max(prev - Math.random() * 5, 40));
    }, 1000);
    return () => clearInterval(interval);
  }, [isPlaying]);

  const handleRun = async () => {
    setIsPlaying(true);
    setRunResult(null);
    setRunError("");
    setActiveTab("console");

    try {
      const result = await executeCode({
        id: `${mission.slug}-${Date.now()}`,
        code: mission.starterCode,
        language: "go",
        timeout: 3,
      });
      setRunResult(result);
    } catch (error) {
      setRunError(error instanceof Error ? error.message : "无法连接到 Gateway 服务");
    } finally {
      setIsPlaying(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col md:flex-row h-full bg-neutral-950 overflow-hidden">
      <aside className="w-full md:w-64 border-r border-neutral-800 bg-neutral-900 flex flex-col shrink-0">
        <div className="p-4 border-b border-neutral-800 flex items-center gap-3 bg-neutral-950/50">
          <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-[#00ADD8] to-blue-600 flex items-center justify-center text-white font-bold shadow-[0_0_15px_rgba(0,173,216,0.3)]">
            Lv.1
          </div>
          <div>
            <div className="text-sm font-bold text-white">Go 实习生</div>
            <div className="text-xs text-neutral-400">经验值: 150/1000</div>
            <div className="h-1.5 w-full bg-neutral-800 rounded-full mt-1.5 overflow-hidden">
              <div className="h-full bg-[#00ADD8] w-[15%]" />
            </div>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto py-4 custom-scrollbar">
          <div className="px-4 mb-2 flex items-center justify-between text-xs font-semibold text-neutral-500 uppercase tracking-wider">
            <span>项目文件</span>
          </div>
          <nav className="space-y-0.5 px-2 mb-6">
            <div className="flex items-center gap-2 px-2 py-1.5 text-sm text-neutral-300 hover:bg-neutral-800 rounded-md cursor-pointer">
              <ChevronRight className="w-4 h-4 text-neutral-500 transition-transform rotate-90" />
              <FolderOpen className="w-4 h-4 text-blue-400" />
              <span>cmd</span>
            </div>
            <div className="flex items-center gap-2 px-2 py-1.5 text-sm text-neutral-300 bg-[#00ADD8]/10 text-[#00ADD8] font-medium rounded-md cursor-pointer ml-4">
              <FileCode2 className="w-4 h-4" />
              <span>main.go</span>
            </div>
            <div className="flex items-center gap-2 px-2 py-1.5 text-sm text-neutral-300 hover:bg-neutral-800 rounded-md cursor-pointer ml-4">
              <FileText className="w-4 h-4 text-neutral-500" />
              <span>go.mod</span>
            </div>
          </nav>

          <div className="px-4 mb-2 text-xs font-semibold text-neutral-500 uppercase tracking-wider">战役任务包</div>
          <nav className="space-y-1 px-2">
            {missions.map((item) => (
              <MissionItem key={item.slug} active={item.slug === mission.slug} mission={item} />
            ))}
          </nav>

          <div className="px-4 mt-8 mb-2 text-xs font-semibold text-neutral-500 uppercase tracking-wider">进阶挑战</div>
          <nav className="space-y-1 px-2">
            <MissionItem mission={{ ...mission, slug: "im-broadcast", title: "高并发 IM 广播", status: "locked" }} icon={<FlaskConical className="w-4 h-4" />} />
            <MissionItem mission={{ ...mission, slug: "flash-sale-rate-limit", title: "双十一抢单限流", status: "locked" }} icon={<LayoutDashboard className="w-4 h-4" />} />
          </nav>
        </div>
      </aside>

      <main className="flex-1 flex flex-col min-w-0 h-[calc(100vh-64px)] md:h-auto relative">
        <header className="h-12 border-b border-neutral-800 bg-neutral-950 flex items-center justify-between px-4 shrink-0 z-10">
          <div className="flex items-center gap-4 min-w-0">
            <div className="flex items-center gap-2 text-sm text-neutral-400 font-mono bg-neutral-900 px-3 py-1 rounded-md border border-neutral-800 border-b-[#00ADD8] shrink-0">
              <FileCode2 className="w-4 h-4 text-[#00ADD8]" />
              main.go
            </div>
            <div className="hidden md:block min-w-0 truncate text-sm text-neutral-300">{mission.title}</div>

            <div className="hidden lg:flex items-center gap-4 text-xs font-mono px-4 border-l border-neutral-800">
              <div className="flex items-center gap-2">
                <Cpu className="w-3.5 h-3.5 text-neutral-500" />
                <span className={`${cpuUsage > 80 ? 'text-red-400' : 'text-neutral-300'}`}>CPU: {cpuUsage.toFixed(0)}%</span>
              </div>
              <div className="flex items-center gap-2">
                <Activity className="w-3.5 h-3.5 text-neutral-500" />
                <span className={`${memUsage > 80 ? 'text-red-400' : 'text-neutral-300'}`}>MEM: {memUsage.toFixed(0)}%</span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button className="p-1.5 text-neutral-400 hover:text-white rounded-md hover:bg-neutral-800 transition-colors">
              <Save className="w-4 h-4" />
            </button>
            <button className="p-1.5 text-neutral-400 hover:text-white rounded-md hover:bg-neutral-800 transition-colors">
              <Settings className="w-4 h-4" />
            </button>
            <button
              onClick={handleRun}
              disabled={isPlaying || mission.status === "locked"}
              className={`flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-semibold transition-all ${isPlaying || mission.status === "locked" ? 'bg-neutral-800 text-neutral-500 cursor-not-allowed' : 'bg-[#00ADD8] text-neutral-950 hover:bg-[#00ADD8]/90 hover:scale-105 shadow-[0_0_15px_rgba(0,173,216,0.3)]'}`}
            >
              {isPlaying ? (
                <div className="w-4 h-4 border-2 border-neutral-500 border-t-transparent rounded-full animate-spin" />
              ) : (
                <Play className="w-4 h-4" />
              )}
              {isPlaying ? '编译运行中...' : mission.status === "locked" ? '任务未解锁' : '运行沙盒'}
            </button>
          </div>
        </header>

        <div className="flex-1 flex flex-col lg:flex-row min-h-0">
          <div className="flex-1 flex flex-col border-r border-neutral-800 min-h-0 relative group bg-[#0d0d0d]">
            <div className="absolute top-2 right-4 opacity-0 group-hover:opacity-100 transition-opacity z-10">
              <span className="text-xs font-mono bg-neutral-800/80 backdrop-blur text-neutral-400 px-2 py-1 rounded">Go 1.22 (Linux/amd64)</span>
            </div>

            <div className="flex-1 overflow-auto flex text-sm font-mono leading-relaxed">
              <div className="py-4 px-3 text-right text-neutral-600 bg-neutral-950 border-r border-neutral-900 select-none">
                {lineNumbers.map(num => (
                  <div key={num}>{num}</div>
                ))}
              </div>
              <pre className="p-4 text-neutral-300 w-full overflow-x-auto whitespace-pre">
                {mission.starterCode}
              </pre>
            </div>
          </div>

          <div className="w-full lg:w-[400px] flex flex-col bg-neutral-900 shrink-0 border-t lg:border-t-0 border-neutral-800 min-h-[350px] lg:min-h-0 relative">
            <div
              className="absolute inset-0 opacity-[0.03] mix-blend-screen pointer-events-none"
              style={{
                backgroundImage: 'url("https://images.unsplash.com/photo-1603943761979-879c839ac8e6?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjb2RlJTIwZWRpdG9yfGVufDF8fHx8MTc3ODA3Nzc5Mnww&ixlib=rb-4.1.0&q=80&w=1080")',
                backgroundSize: 'cover',
                backgroundPosition: 'center',
              }}
            />

            <div className="flex items-center border-b border-neutral-800 px-2 bg-neutral-950 z-10">
              <TabButton
                active={activeTab === "code"}
                onClick={() => setActiveTab("code")}
                icon={<Code2 className="w-4 h-4" />}
                label="说明"
              />
              <TabButton
                active={activeTab === "console"}
                onClick={() => setActiveTab("console")}
                icon={<Terminal className="w-4 h-4" />}
                label="控制台"
              />
              <TabButton
                active={activeTab === "ai"}
                onClick={() => setActiveTab("ai")}
                icon={<Bot className={`w-4 h-4 ${activeTab === 'ai' ? 'text-purple-400' : 'text-neutral-400'}`} />}
                label="AI CTO"
                badge={!isPlaying && activeTab !== 'ai' ? "1" : undefined}
              />
            </div>

            <div className="flex-1 overflow-auto p-4 z-10">
              {activeTab === "code" && (
                <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="space-y-4 text-sm text-neutral-300">
                  <h3 className="font-bold text-lg text-white flex items-center gap-2">
                    <AlertTriangle className="w-5 h-5 text-yellow-500" />
                    任务目标
                  </h3>
                  <p>{mission.background[0]}</p>
                  <div className="bg-neutral-950 p-4 rounded-xl border border-neutral-800 relative overflow-hidden">
                    <div className="absolute top-0 left-0 w-1 h-full bg-red-500"></div>
                    <p className="font-semibold text-red-400 mb-2 flex items-center gap-2">
                      <Activity className="w-4 h-4" /> 修复目标：
                    </p>
                    <ul className="list-disc pl-5 space-y-1">
                      {mission.objectives.map((objective) => (
                        <li key={objective}>{objective}</li>
                      ))}
                    </ul>
                  </div>
                  <div className="p-3 bg-blue-950/20 border border-blue-900/30 rounded-lg">
                    <p className="text-blue-400 text-xs flex items-start gap-2">
                      <Bot className="w-4 h-4 shrink-0" />
                      <span><strong>提示：</strong>{mission.hints[0]}</span>
                    </p>
                  </div>
                </motion.div>
              )}

              {activeTab === "console" && (
                <ConsolePanel isPlaying={isPlaying} result={runResult} error={runError} />
              )}

              {activeTab === "ai" && (
                <div className="flex flex-col h-full">
                  <div className="flex-1 space-y-4 pb-4">
                    <AnimatePresence>
                      <motion.div
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        className="flex gap-3"
                      >
                        <div className="w-8 h-8 rounded-full bg-purple-500/20 flex items-center justify-center shrink-0 border border-purple-500/50">
                          <Bot className="w-5 h-5 text-purple-400" />
                        </div>
                        <div className="bg-neutral-800 p-4 rounded-2xl rounded-tl-none text-sm text-neutral-200 shadow-lg">
                          <p className="mb-3 leading-relaxed">实习生，这里会根据沙盒输出给出代码诊断。当前 MVP 先保留静态导师提示。</p>
                          <p className="mb-2 font-semibold text-white">本关重点：</p>
                          <ul className="list-disc pl-4 space-y-2 text-neutral-300 mb-4">
                            {mission.hints.map((hint) => (
                              <li key={hint}>{hint}</li>
                            ))}
                          </ul>
                          <div className="bg-neutral-950 p-3 rounded-lg border border-neutral-700 font-mono text-xs">
                            <span className="text-green-400 font-bold mb-1 block">// 下一步：</span>
                            <span className="text-neutral-400">点击“运行沙盒”查看真实执行结果，再根据错误信息迭代代码。</span>
                          </div>
                        </div>
                      </motion.div>
                    </AnimatePresence>
                    <div ref={endOfMessagesRef} />
                  </div>

                  <div className="mt-auto relative pt-4 border-t border-neutral-800">
                    <input
                      type="text"
                      value={chatInput}
                      onChange={(e) => setChatInput(e.target.value)}
                      placeholder="向 AI CTO 提问，例如 '如何正确使用 copy()?'"
                      className="w-full bg-neutral-950 border border-neutral-800 rounded-xl py-3 pl-4 pr-12 text-sm text-white focus:outline-none focus:border-purple-500/50 focus:ring-1 focus:ring-purple-500/50 transition-all placeholder:text-neutral-600"
                    />
                    <button
                      className="absolute right-2 top-1/2 -translate-y-1/2 mt-2 p-2 text-neutral-400 hover:text-purple-400 hover:bg-purple-400/10 rounded-lg transition-colors"
                      disabled={!chatInput.trim()}
                    >
                      <Send className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

function MissionItem({ active, mission, icon }: { active?: boolean, mission: Mission, icon?: React.ReactNode }) {
  const className = `w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm transition-all text-left ${active ? 'bg-[#00ADD8]/10 text-[#00ADD8] font-medium border border-[#00ADD8]/20' : 'text-neutral-400 hover:bg-neutral-800 border border-transparent'}`;
  const content = (
    <>
      <div className="flex items-center gap-2.5 min-w-0">
        {icon || <Code2 className={`w-4 h-4 shrink-0 ${active ? 'text-[#00ADD8]' : 'text-neutral-500'}`} />}
        <span className="truncate">{mission.title}</span>
      </div>
      {mission.status === 'locked' && <div className="w-1.5 h-1.5 rounded-full bg-neutral-700 shrink-0"></div>}
      {mission.status === 'in-progress' && <div className="w-1.5 h-1.5 rounded-full bg-yellow-500 shadow-[0_0_8px_#eab308] animate-pulse shrink-0"></div>}
      {mission.status === 'completed' && <div className="w-1.5 h-1.5 rounded-full bg-green-500 shadow-[0_0_8px_#22c55e] shrink-0"></div>}
    </>
  );

  if (mission.status === 'locked') {
    return <button className={className}>{content}</button>;
  }

  return <Link to={`/dashboard?mission=${mission.slug}`} className={className}>{content}</Link>;
}

function ConsolePanel({ isPlaying, result, error }: { isPlaying: boolean; result: SandboxResponse | null; error: string }) {
  if (isPlaying) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="font-mono text-sm space-y-2 text-neutral-300">
        <span className="text-[#00ADD8] font-bold">$</span> go run main.go
        <div className="mt-2 text-neutral-500 flex items-center gap-2">
          <div className="w-2 h-2 rounded-full bg-yellow-500 animate-ping" />
          Compiling and linking...
        </div>
      </motion.div>
    );
  }

  if (error) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="font-mono text-sm space-y-3 text-neutral-300">
        <div className="text-white"><span className="text-[#00ADD8] font-bold">$</span> go run main.go</div>
        <div className="rounded-lg border border-red-900 bg-red-950 p-3 text-red-300">{error}</div>
      </motion.div>
    );
  }

  if (!result) {
    return (
      <div className="font-mono text-sm text-neutral-500">
        点击“运行沙盒”后，这里会显示 Gateway 返回的真实执行结果。
      </div>
    );
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="font-mono text-sm space-y-3 text-neutral-300">
      <div className="text-white"><span className="text-[#00ADD8] font-bold">$</span> go run main.go</div>
      <div className="grid grid-cols-2 gap-2 text-xs text-neutral-400">
        <Metric label="status" value={result.status} tone={result.status === "success" ? "text-green-400" : result.status === "timeout" ? "text-yellow-400" : "text-red-400"} />
        <Metric label="exit_code" value={String(result.exit_code)} />
        <Metric label="duration" value={`${(result.duration / 1_000_000).toFixed(1)}ms`} />
        <Metric label="id" value={result.id.split("-").slice(0, -1).join("-")} />
      </div>
      {result.stdout && (
        <pre className="whitespace-pre-wrap rounded-lg border border-green-900/50 bg-green-950/30 p-3 text-green-300">{result.stdout}</pre>
      )}
      {result.stderr && (
        <pre className="whitespace-pre-wrap rounded-lg border border-red-900/50 bg-red-950/50 p-3 text-red-300">{result.stderr}</pre>
      )}
    </motion.div>
  );
}

function Metric({ label, value, tone = "text-neutral-300" }: { label: string; value: string; tone?: string }) {
  return (
    <div className="rounded border border-neutral-800 bg-neutral-950 p-2">
      <div className="mb-1 text-[10px] uppercase tracking-wide text-neutral-600">{label}</div>
      <div className={`truncate ${tone}`}>{value}</div>
    </div>
  );
}

function TabButton({ active, label, icon, badge, onClick }: { active: boolean, label: string, icon: React.ReactNode, badge?: string, onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`relative flex items-center gap-2 px-4 py-3.5 text-sm font-medium transition-colors ${active ? 'text-white' : 'text-neutral-500 hover:text-neutral-300'}`}
    >
      {icon}
      {label}
      {badge && (
        <span className="absolute top-2 right-1.5 bg-red-500 text-white text-[10px] w-4 h-4 flex items-center justify-center rounded-full animate-bounce">
          {badge}
        </span>
      )}
      {active && (
        <motion.div
          layoutId="activeTab"
          className="absolute bottom-0 left-0 right-0 h-0.5 bg-[#00ADD8] shadow-[0_0_10px_rgba(0,173,216,0.8)]"
        />
      )}
    </button>
  );
}
