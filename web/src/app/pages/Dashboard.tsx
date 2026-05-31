import { Link, useSearchParams } from "react-router";
import { useEffect, useMemo, useState } from "react";
import { ArrowRight, BookOpen, CheckCircle2, Clock3, Code2, Loader2, Play, RotateCcw, Terminal } from "lucide-react";
import { executeCode } from "../../api/execute";
import type { SandboxResponse } from "../../api/types";
import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Progress } from "../components/ui/progress";
import { Separator } from "../components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../components/ui/tabs";
import { goBasicsChapters } from "../data/goBasicsCourse";
import { getMissionBySlug, missions, statusMeta } from "../data/missions";

export function Dashboard() {
  const [searchParams] = useSearchParams();
  const mission = getMissionBySlug(searchParams.get("mission"));
  const [isRunning, setIsRunning] = useState(false);
  const [runResult, setRunResult] = useState<SandboxResponse | null>(null);
  const [runError, setRunError] = useState("");

  useEffect(() => {
    setRunResult(null);
    setRunError("");
  }, [mission.slug]);

  const nextChapter = goBasicsChapters[0];
  const progress = useMemo(() => Math.round((1 / Math.max(goBasicsChapters.length + missions.length, 1)) * 100), []);

  const handleRun = async () => {
    setIsRunning(true);
    setRunResult(null);
    setRunError("");

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
      setIsRunning(false);
    }
  };

  return (
    <main className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-6 md:px-6 md:py-8">
      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.25fr)_minmax(320px,0.75fr)]">
        <Card className="border-primary/20 bg-background shadow-sm">
          <CardHeader className="gap-4">
            <div className="flex flex-wrap items-center gap-2">
              <Badge>访客演示</Badge>
              <Badge variant="secondary">本地会话</Badge>
              <Badge variant="outline">source: derivedFromContent</Badge>
            </div>
            <div className="grid gap-6 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
              <div>
                <CardTitle className="text-3xl md:text-4xl">学习总览</CardTitle>
                <CardDescription className="mt-3 max-w-2xl text-base leading-7">
                  当前页面只展示本地会话和内容派生状态，不代表账号级真实进度。下一步建议会优先引导你完成 Go 基础和第一条实习任务线。
                </CardDescription>
              </div>
              <Button asChild>
                <Link to={`/courses/go-basics/${nextChapter.slug}`}>
                  继续学习
                  <ArrowRight data-icon="inline-end" />
                </Link>
              </Button>
            </div>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-3">
            <MetricCard label="学习路径" value="Go 基础" helper="source: derivedFromContent" />
            <MetricCard label="当前任务" value={`Day ${mission.day}`} helper="source: localSession" />
            <MetricCard label="演示进度" value={`${progress}%`} helper="source: staticMock" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock3 className="text-primary" />
              今日建议
            </CardTitle>
            <CardDescription>按顺序完成即可形成学习闭环。</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            <ChecklistItem done text="阅读 Go 基础第 1 章" />
            <ChecklistItem text="运行 1 次章节 sandbox" />
            <ChecklistItem text={`挑战任务：${mission.title}`} />
          </CardContent>
        </Card>
      </section>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BookOpen className="text-primary" />
              下一步学习
            </CardTitle>
            <CardDescription>从课程阅读进入动手任务。</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <div className="rounded-2xl border bg-muted/40 p-4">
              <div className="mb-2 flex flex-wrap items-center gap-2">
                <Badge variant="outline">Go 基础训练营</Badge>
                <Badge variant="secondary">第 {nextChapter.order} 章</Badge>
              </div>
              <h2 className="text-xl font-semibold">{nextChapter.title}</h2>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">{nextChapter.summary}</p>
              <div className="mt-4 flex flex-wrap gap-2">
                <Button asChild>
                  <Link to={`/courses/go-basics/${nextChapter.slug}`}>阅读章节</Link>
                </Button>
                <Button asChild variant="outline">
                  <Link to={`/courses/go-basics/${nextChapter.slug}#exercise`}>直接运行练习</Link>
                </Button>
              </div>
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
              <PathSummary title="Go 基础" value={`${goBasicsChapters.length} 章`} />
              <PathSummary title="后端实习" value={`${missions.length} 个任务`} />
              <PathSummary title="工程进阶" value="即将开放" muted />
              <PathSummary title="AI 全栈" value="即将开放" muted />
            </div>
          </CardContent>
        </Card>

        <Card id="sandbox">
          <CardHeader>
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Terminal className="text-primary" />
                  沙盒快捷运行
                </CardTitle>
                <CardDescription>source: localSession · 运行结果刷新后可丢失。</CardDescription>
              </div>
              <Button onClick={handleRun} disabled={isRunning || mission.status === "locked"}>
                {isRunning ? <Loader2 className="animate-spin" data-icon="inline-start" /> : <Play data-icon="inline-start" />}
                {isRunning ? "运行中" : mission.status === "locked" ? "任务未解锁" : "运行当前任务"}
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="brief">
              <TabsList>
                <TabsTrigger value="brief">任务说明</TabsTrigger>
                <TabsTrigger value="code">代码</TabsTrigger>
                <TabsTrigger value="console">控制台</TabsTrigger>
              </TabsList>
              <TabsContent value="brief" className="mt-4">
                <div className="rounded-2xl border bg-muted/40 p-4">
                  <div className="mb-3 flex flex-wrap items-center gap-2">
                    <Badge variant="outline">Day {mission.day}</Badge>
                    <Badge className={statusMeta[mission.status].className}>{statusMeta[mission.status].label}</Badge>
                  </div>
                  <h3 className="font-semibold">{mission.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">{mission.background[0]}</p>
                </div>
              </TabsContent>
              <TabsContent value="code" className="mt-4">
                <CodeBlock code={mission.starterCode} />
              </TabsContent>
              <TabsContent value="console" className="mt-4">
                <ConsolePanel isRunning={isRunning} result={runResult} error={runError} />
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </section>

      <section className="grid gap-4 md:grid-cols-3">
        {missions.map((item) => (
          <Card key={item.slug}>
            <CardHeader>
              <div className="flex items-center justify-between gap-3">
                <Badge variant="outline">Day {item.day}</Badge>
                <Badge className={statusMeta[item.status].className}>{statusMeta[item.status].label}</Badge>
              </div>
              <CardTitle className="text-lg">{item.title}</CardTitle>
              <CardDescription>{item.chapter} · {item.duration}</CardDescription>
            </CardHeader>
            <CardContent>
              {item.status === "locked" ? (
                <Button variant="secondary" className="w-full" disabled>
                  查看条件
                </Button>
              ) : (
                <Button asChild variant="outline" className="w-full">
                  <Link to={`/missions/${item.slug}`}>查看任务</Link>
                </Button>
              )}
            </CardContent>
          </Card>
        ))}
      </section>
    </main>
  );
}

function MetricCard({ label, value, helper }: { label: string; value: string; helper: string }) {
  return (
    <div className="rounded-2xl border bg-muted/30 p-4">
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className="mt-2 text-2xl font-bold">{value}</div>
      <div className="mt-3 text-xs text-muted-foreground">{helper}</div>
    </div>
  );
}

function ChecklistItem({ done, text }: { done?: boolean; text: string }) {
  return (
    <div className="flex items-center gap-3 rounded-xl border bg-background p-3 text-sm">
      <CheckCircle2 className={done ? "text-primary" : "text-muted-foreground"} />
      <span>{text}</span>
    </div>
  );
}

function PathSummary({ title, value, muted }: { title: string; value: string; muted?: boolean }) {
  return (
    <div className="rounded-xl border bg-background p-3">
      <div className="text-sm font-medium">{title}</div>
      <div className={muted ? "mt-1 text-sm text-muted-foreground" : "mt-1 text-sm text-primary"}>{value}</div>
      {!muted && <Progress value={18} className="mt-3" />}
    </div>
  );
}

function CodeBlock({ code }: { code: string }) {
  return (
    <pre className="max-h-[420px] overflow-auto rounded-2xl border bg-slate-950 p-4 text-sm leading-6 text-slate-100">
      <code>{code}</code>
    </pre>
  );
}

function ConsolePanel({ isRunning, result, error }: { isRunning: boolean; result: SandboxResponse | null; error: string }) {
  if (isRunning) {
    return (
      <div className="rounded-2xl border bg-slate-950 p-4 font-mono text-sm text-slate-200">
        <div className="flex items-center gap-2 text-cyan-300">
          <Loader2 className="animate-spin" />
          go run main.go
        </div>
        <div className="mt-3 text-slate-500">Compiling and linking...</div>
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <Terminal />
        <AlertTitle>无法连接到代码运行服务</AlertTitle>
        <AlertDescription>{error}。请确认本地 Gateway 和 Sandbox Engine 已启动。</AlertDescription>
      </Alert>
    );
  }

  if (!result) {
    return (
      <div className="rounded-2xl border border-dashed bg-muted/30 p-6 text-sm text-muted-foreground">
        点击“运行当前任务”后，这里会显示当前浏览器会话内的真实执行结果。
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3 rounded-2xl border bg-slate-950 p-4 font-mono text-sm text-slate-200">
      <div className="grid gap-2 sm:grid-cols-3">
        <ConsoleMetric label="status" value={result.status} />
        <ConsoleMetric label="exit" value={String(result.exit_code)} />
        <ConsoleMetric label="duration" value={`${result.duration}ms`} />
      </div>
      <TerminalOutput label="stdout" value={result.stdout} />
      <TerminalOutput label="stderr" value={result.stderr} />
      <Button variant="secondary" size="sm" onClick={() => window.location.reload()}>
        <RotateCcw data-icon="inline-start" />
        清空本地会话展示
      </Button>
    </div>
  );
}

function ConsoleMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-slate-800 bg-slate-900 p-3">
      <div className="text-xs uppercase text-slate-500">{label}</div>
      <div className="mt-1 truncate text-slate-100">{value}</div>
    </div>
  );
}

function TerminalOutput({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="mb-1 text-xs uppercase text-slate-500">{label}</div>
      <pre className="min-h-12 whitespace-pre-wrap rounded-xl border border-slate-800 bg-black p-3 text-slate-100">
        {value || <span className="text-slate-600">无输出</span>}
      </pre>
    </div>
  );
}
