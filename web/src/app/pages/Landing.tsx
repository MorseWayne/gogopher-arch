import { motion } from "motion/react";
import { Link } from "react-router";
import { ArrowRight, Shield, Beaker, Zap, Bot, Map, ChevronRight, CheckCircle2, Circle, Activity, BookOpen } from "lucide-react";

export function Landing() {
  return (
    <div className="flex flex-col w-full">
      {/* Hero Section */}
      <section className="relative pt-32 pb-20 md:pt-48 md:pb-32 overflow-hidden">
        <div 
          className="absolute inset-0 opacity-10 mix-blend-overlay pointer-events-none"
          style={{
            backgroundImage: 'url("https://images.unsplash.com/photo-1601042879364-f3947d3f9c16?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjeWJlcnB1bmslMjBjaXR5fGVufDF8fHx8MTc3ODA0NjM4M3ww&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral")',
            backgroundSize: 'cover',
            backgroundPosition: 'center',
          }}
        />
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-[#00ADD8]/20 via-neutral-950 to-neutral-950"></div>
        <div className="container mx-auto px-6 relative z-10">
          <div className="max-w-4xl mx-auto text-center">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
              className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-neutral-900 border border-neutral-800 text-sm text-neutral-300 mb-8"
            >
              <span className="flex h-2 w-2 rounded-full bg-[#00ADD8]"></span>
              v1.0 MVP 闭环开发中
            </motion.div>
            
            <motion.h1 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="text-5xl md:text-7xl font-extrabold tracking-tight mb-8"
            >
              GoGopher Arch
              <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#00ADD8] to-blue-500">
                架构师进化之路
              </span>
            </motion.h1>

            <motion.p 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="text-xl text-neutral-400 mb-12 max-w-2xl mx-auto leading-relaxed"
            >
              一个沉浸式、由浅入深、强互动的 Go 语言全栈进阶学习平台。拒绝枯燥文档灌输，通过虚拟职场实战和可视化反馈，带你从 Go 语言实习生进化为资深架构师。
            </motion.p>

            <motion.div 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.3 }}
              className="flex flex-col sm:flex-row items-center justify-center gap-4"
            >
              <Link
                to="/courses/go-basics"
                className="w-full sm:w-auto px-8 py-4 bg-[#00ADD8] text-neutral-950 font-bold rounded-xl hover:bg-[#00ADD8]/90 transition-colors flex items-center justify-center gap-2 text-lg group"
              >
                开始 Go 基础训练营
                <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
              </Link>
              <Link
                to="/dashboard"
                className="w-full sm:w-auto px-8 py-4 bg-neutral-900 text-white font-medium rounded-xl hover:bg-neutral-800 border border-neutral-800 transition-colors flex items-center justify-center gap-2 text-lg"
              >
                进入沙盒体验
              </Link>
            </motion.div>
          </div>
        </div>

        {/* Badges/Stack */}
        <div className="mt-20 border-y border-neutral-900 bg-neutral-950/50 backdrop-blur-sm py-8">
          <div className="container mx-auto px-6 flex flex-wrap justify-center gap-8 md:gap-16 opacity-60">
            <div className="flex items-center gap-2 font-mono text-lg font-bold">
              <span className="text-[#00ADD8]">Go</span> 1.22+
            </div>
            <div className="flex items-center gap-2 font-mono text-lg font-bold">
              <span className="text-[#61DAFB]">React</span>
            </div>
            <div className="flex items-center gap-2 font-mono text-lg font-bold">
              <span className="text-blue-400">Docker</span> Sandbox
            </div>
            <div className="flex items-center gap-2 font-mono text-lg font-bold">
              <span className="text-purple-400">Gemini</span> AI
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-24 bg-neutral-950">
        <div className="container mx-auto px-6">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-5xl font-bold mb-4">核心特性</h2>
            <p className="text-neutral-400 text-lg max-w-2xl mx-auto">
              打破传统学习方式，通过游戏化和可视化技术重塑学习体验
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
            <FeatureCard
              icon={<BookOpen className="w-8 h-8 text-[#00ADD8]" />}
              title="Go 基础训练营"
              description="参考并改编《Go 语言圣经中文版》，按 13 章重写为可学习、可运行、可衔接实习任务的基础课程。"
            />
            <FeatureCard
              icon={<Shield className="w-8 h-8 text-yellow-500" />}
              title="战役模式 (RPG Style)"
              description="扮演职场角色，通过解决“双十一抢单限流”、“高并发 IM 系统崩溃”等真实业务挑战来晋升。"
            />
            <FeatureCard 
              icon={<Beaker className="w-8 h-8 text-green-500" />}
              title="交互式沙盒 (Sandbox)"
              description="内置 Cloud IDE 和模拟集群压测环境，即写即练，直观观察系统吞吐量与 P99 延迟。"
            />
            <FeatureCard 
              icon={<Zap className="w-8 h-8 text-red-500" />}
              title="可视化“大爆炸”"
              description="当代码导致协程爆炸或内存泄露时，通过动态动画直观展示 Go Runtime 内部故障过程。"
            />
            <FeatureCard 
              icon={<Bot className="w-8 h-8 text-blue-500" />}
              title="AI CTO 评审"
              description="集成大语言模型，模仿专家口吻进行 Code Review，指出并发陷阱并推荐架构优化方案。"
            />
            <FeatureCard 
              icon={<Map className="w-8 h-8 text-purple-500" />}
              title="全栈地图"
              description="覆盖基础语法、高性能网络编程、云原生 (K8s/gRPC)、区块链及 Go Runtime 深度调优。"
            />
          </div>
        </div>
      </section>

      {/* Immersive Preview Section */}
      <section className="py-24 bg-neutral-900 border-y border-neutral-800 relative overflow-hidden">
        <div 
          className="absolute inset-0 opacity-[0.03] mix-blend-screen pointer-events-none"
          style={{
            backgroundImage: 'url("https://images.unsplash.com/photo-1558494949-ef010cbdcc31?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxzZXJ2ZXIlMjByb29tfGVufDF8fHx8MTc3Nzk5Mjk5OHww&ixlib=rb-4.1.0&q=80&w=1080")',
            backgroundSize: 'cover',
            backgroundPosition: 'center',
          }}
        />
        <div className="container mx-auto px-6 relative z-10">
          <div className="flex flex-col lg:flex-row items-center gap-16">
            <div className="flex-1 space-y-8">
              <h2 className="text-3xl md:text-5xl font-bold leading-tight">
                不止于代码，<br />
                <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-[#00ADD8]">洞察 Runtime 级真相</span>
              </h2>
              <p className="text-lg text-neutral-400 leading-relaxed">
                在真实的微服务压测环境中，你的每一行代码都可能引发雪崩。我们的内置沙盒不仅能运行代码，更能透视 CPU、内存和 Goroutine 的实时状态。当故障发生时，AI CTO 将立刻介入，带你复盘灾难现场。
              </p>
              <ul className="space-y-4">
                <li className="flex items-center gap-3 text-neutral-300 font-medium">
                  <div className="w-8 h-8 rounded-full bg-[#00ADD8]/10 flex items-center justify-center text-[#00ADD8]">
                    <Activity className="w-4 h-4" />
                  </div>
                  真实容器隔离，毫秒级指标监控
                </li>
                <li className="flex items-center gap-3 text-neutral-300 font-medium">
                  <div className="w-8 h-8 rounded-full bg-purple-500/10 flex items-center justify-center text-purple-400">
                    <Bot className="w-4 h-4" />
                  </div>
                  Gemini 驱动的资深架构师 Code Review
                </li>
              </ul>
            </div>
            
            {/* Mock Dashboard Window */}
            <div className="flex-1 w-full">
              <div className="bg-neutral-950 border border-neutral-800 rounded-2xl overflow-hidden shadow-2xl shadow-[#00ADD8]/10 transform perspective-1000 rotate-y-[-5deg] rotate-x-[2deg]">
                <div className="h-10 bg-neutral-900 border-b border-neutral-800 flex items-center px-4 gap-2">
                  <div className="w-3 h-3 rounded-full bg-red-500/80"></div>
                  <div className="w-3 h-3 rounded-full bg-yellow-500/80"></div>
                  <div className="w-3 h-3 rounded-full bg-green-500/80"></div>
                </div>
                <div className="p-6 font-mono text-sm leading-relaxed">
                  <div className="text-[#00ADD8]">$ go test -bench=. -benchmem</div>
                  <div className="text-neutral-500 mt-2">goos: linux</div>
                  <div className="text-neutral-500">goarch: amd64</div>
                  <div className="text-neutral-500 mb-4">pkg: github.com/gogopher/im</div>
                  <div className="flex justify-between text-neutral-300 border-b border-neutral-800 pb-2 mb-2">
                    <span>BenchmarkBroadcast-8</span>
                    <span>100000</span>
                    <span>14562 ns/op</span>
                    <span className="text-red-400">4502 B/op</span>
                  </div>
                  <motion.div 
                    initial={{ opacity: 0 }}
                    whileInView={{ opacity: 1 }}
                    viewport={{ once: true }}
                    className="mt-6 p-4 bg-purple-950/30 border border-purple-900/50 rounded-xl"
                  >
                    <div className="flex gap-3">
                      <Bot className="w-5 h-5 text-purple-400 shrink-0" />
                      <div className="text-purple-200">
                        <span className="font-bold text-purple-400">AI CTO: </span>
                        每次广播分配了 4.5KB 内存，在高并发下 GC 压力极大。建议引入 <code>sync.Pool</code> 复用 Message 对象。
                      </div>
                    </div>
                  </motion.div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Roadmap Section */}
      <section id="roadmap" className="py-24 bg-neutral-900 border-y border-neutral-800">
        <div className="container mx-auto px-6">
          <div className="max-w-3xl mx-auto">
            <h2 className="text-3xl md:text-5xl font-bold mb-16 text-center">路线图 (Roadmap)</h2>
            
            <div className="space-y-12 relative before:absolute before:inset-0 before:ml-5 before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-neutral-800 before:to-transparent">
              
              {/* Stage 1 */}
              <div className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
                <div className="flex items-center justify-center w-10 h-10 rounded-full border-4 border-neutral-900 bg-[#00ADD8] text-white shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 shadow-[0_0_0_4px_rgba(0,173,216,0.2)]">
                  <span className="font-bold">1</span>
                </div>
                <div className="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] bg-neutral-950 p-6 rounded-2xl border border-neutral-800 hover:border-[#00ADD8]/50 transition-colors">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-bold text-xl text-white">第一阶段：MVP 闭环</h3>
                    <span className="text-xs font-semibold px-3 py-1 rounded-full bg-[#00ADD8]/10 text-[#00ADD8] border border-[#00ADD8]/20 animate-pulse">正在进行</span>
                  </div>
                  <ul className="space-y-3">
                    <RoadmapItem text="项目立项与设计文档确认" done={true} />
                    <RoadmapItem text="Go Monorepo 基础脚手架搭建" done={false} />
                    <RoadmapItem text="基于 Docker 的安全代码运行沙盒 (Sandbox)" done={false} />
                    <RoadmapItem text="Lv.1 实习生基础任务包 (Slice/Map/Defer 陷阱)" done={false} />
                  </ul>
                </div>
              </div>

              {/* Stage 2 */}
              <div className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group">
                <div className="flex items-center justify-center w-10 h-10 rounded-full border-4 border-neutral-900 bg-neutral-800 text-neutral-400 shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
                  <span className="font-bold">2</span>
                </div>
                <div className="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] bg-neutral-950/50 p-6 rounded-2xl border border-neutral-800/50">
                  <h3 className="font-bold text-xl text-white mb-4">第二阶段：深度反馈</h3>
                  <ul className="space-y-3">
                    <RoadmapItem text="实时监控指标可视化 (Goroutine/Memory)" done={false} />
                    <RoadmapItem text="“大爆炸”崩溃动画引擎" done={false} />
                    <RoadmapItem text="AI CTO 初版接入" done={false} />
                  </ul>
                </div>
              </div>

              {/* Stage 3 */}
              <div className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group">
                <div className="flex items-center justify-center w-10 h-10 rounded-full border-4 border-neutral-900 bg-neutral-800 text-neutral-400 shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
                  <span className="font-bold">3</span>
                </div>
                <div className="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] bg-neutral-950/50 p-6 rounded-2xl border border-neutral-800/50">
                  <h3 className="font-bold text-xl text-white mb-4">第三阶段：全栈进阶</h3>
                  <ul className="space-y-3">
                    <RoadmapItem text="高性能网络编程实战 (IM 系统)" done={false} />
                    <RoadmapItem text="云原生微服务治理与区块链实战" done={false} />
                  </ul>
                </div>
              </div>

            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-24 bg-neutral-950 text-center relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-t from-[#00ADD8]/10 to-transparent"></div>
        <div className="container mx-auto px-6 relative z-10">
          <h2 className="text-4xl md:text-5xl font-bold mb-6">准备好开始进化了吗？</h2>
          <p className="text-xl text-neutral-400 mb-10 max-w-2xl mx-auto">
            加入我们，在真实的业务场景中锤炼你的架构能力。
          </p>
          <div className="bg-neutral-900 border border-neutral-800 p-4 rounded-xl inline-flex flex-col items-start gap-2 mb-8">
            <span className="text-neutral-500 text-sm font-mono">快速克隆项目</span>
            <code className="text-[#00ADD8] font-mono text-sm sm:text-base">git clone https://github.com/MorseWayne/gogopher-arch.git</code>
          </div>
        </div>
      </section>
    </div>
  );
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode, title: string, description: string }) {
  return (
    <div className="bg-neutral-900/50 border border-neutral-800 p-8 rounded-2xl hover:bg-neutral-900 transition-colors">
      <div className="mb-6 p-4 bg-neutral-950 rounded-xl inline-block border border-neutral-800">
        {icon}
      </div>
      <h3 className="text-xl font-bold mb-3 text-white">{title}</h3>
      <p className="text-neutral-400 leading-relaxed">{description}</p>
    </div>
  );
}

function RoadmapItem({ text, done }: { text: string, done: boolean }) {
  return (
    <li className="flex items-start gap-3 text-sm md:text-base">
      {done ? (
        <CheckCircle2 className="w-5 h-5 text-[#00ADD8] shrink-0 mt-0.5" />
      ) : (
        <Circle className="w-5 h-5 text-neutral-600 shrink-0 mt-0.5" />
      )}
      <span className={done ? "text-neutral-300 line-through decoration-neutral-600 opacity-80" : "text-neutral-300"}>
        {text}
      </span>
    </li>
  );
}
