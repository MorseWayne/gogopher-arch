import type { ReactNode } from "react";
import { Link, useParams } from "react-router";
import { ArrowLeft, ArrowRight, BookOpen, CheckCircle2, CheckSquare, Clock3, Code2, GraduationCap, Lightbulb, Lock, MapPin, Square, Star, Target, Terminal } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Separator } from "../components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../components/ui/tabs";
import { goFunctions, goKeywords, missionCatalog, missions, statusMeta, type Mission } from "../data/missions";

export function MissionDetail() {
  const { slug } = useParams();
  const requestedMission = slug ? missionCatalog[slug] : undefined;
  const mission = requestedMission ?? missions[0];
  const missionIndex = missions.findIndex((item) => item.slug === mission.slug);
  const previousMission = missions[missionIndex - 1];
  const nextMission = missions[missionIndex + 1];
  const completedCriteria = mission.status === "completed" ? mission.criteria.length : mission.status === "in-progress" ? Math.min(2, mission.criteria.length) : 0;

  return (
    <main className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-6 md:px-6 md:py-8">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Button asChild variant="ghost" size="sm">
          <Link to="/dashboard">
            <ArrowLeft data-icon="inline-start" />
            返回学习总览
          </Link>
        </Button>
        <div className="flex items-center gap-2">
          <MissionNavButton mission={previousMission} label="上一关" direction="previous" />
          <MissionNavButton mission={nextMission} label="下一关" direction="next" />
        </div>
      </div>

      {!requestedMission && slug && (
        <Alert>
          <MapPin />
          <AlertTitle>未找到指定任务</AlertTitle>
          <AlertDescription>已展示当前可继续的任务。你也可以返回学习总览选择其他任务。</AlertDescription>
        </Alert>
      )}

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
        <Card className="border-primary/20">
          <CardHeader className="gap-4">
            <div className="flex flex-wrap items-center gap-2">
              <Badge>Day {mission.day}</Badge>
              <Badge variant="secondary">{mission.chapter}</Badge>
              <Badge className={statusMeta[mission.status].className}>{statusMeta[mission.status].label}</Badge>
            </div>
            <div>
              <CardTitle className="text-3xl md:text-5xl">{mission.title}</CardTitle>
              <CardDescription className="mt-4 max-w-3xl text-base leading-7">{mission.background[0]}</CardDescription>
            </div>
            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <MissionMeta icon={<Clock3 />} text={mission.duration} />
              <MissionMeta icon={<Star />} text={mission.difficulty} />
              <MissionMeta icon={<MapPin />} text={`前置: ${mission.prerequisite.join("、")}`} />
            </div>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">任务状态</CardTitle>
            <CardDescription>本轮只展示本地任务状态，不做账号级持久化。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <MissionCta mission={mission} />
            <Separator />
            <div className="text-sm text-muted-foreground">完成验收项 {completedCriteria}/{mission.criteria.length}</div>
            <div className="h-2 rounded-full bg-muted">
              <div className="h-2 rounded-full bg-primary" style={{ width: `${mission.criteria.length ? (completedCriteria / mission.criteria.length) * 100 : 0}%` }} />
            </div>
          </CardContent>
        </Card>
      </section>

      <section className="grid gap-6 xl:grid-cols-[minmax(0,0.95fr)_minmax(420px,1.05fr)]">
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BookOpen className="text-primary" />
                任务 Brief
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Tabs defaultValue="background">
                <TabsList>
                  <TabsTrigger value="background">背景</TabsTrigger>
                  <TabsTrigger value="objectives">目标</TabsTrigger>
                  <TabsTrigger value="hints">小课</TabsTrigger>
                  <TabsTrigger value="knowledge">知识</TabsTrigger>
                </TabsList>
                <TabsContent value="background" className="mt-4 space-y-3 text-sm leading-7 text-muted-foreground">
                  {mission.background.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
                </TabsContent>
                <TabsContent value="objectives" className="mt-4">
                  <ListBlock items={mission.objectives} icon={<Target />} />
                </TabsContent>
                <TabsContent value="hints" className="mt-4">
                  <Alert>
                    <Lightbulb />
                    <AlertTitle>任务前小课</AlertTitle>
                    <AlertDescription>
                      <ul className="mt-2 list-disc pl-5">
                        {mission.hints.map((hint) => <li key={hint}>{hint}</li>)}
                      </ul>
                    </AlertDescription>
                  </Alert>
                </TabsContent>
                <TabsContent value="knowledge" className="mt-4">
                  <div className="flex flex-wrap gap-2">
                    {mission.knowledge.map((item) => <Badge key={item} variant="secondary">{item}</Badge>)}
                  </div>
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">验收标准</CardTitle>
              <CardDescription>验收勾选只表示本地展示状态，不做持久化。</CardDescription>
            </CardHeader>
            <CardContent>
              {mission.criteria.length === 0 ? (
                <Alert>
                  <AlertTitle>暂无验收标准</AlertTitle>
                  <AlertDescription>请先阅读任务 Brief。</AlertDescription>
                </Alert>
              ) : (
                <ul className="space-y-3">
                  {mission.criteria.map((criterion, index) => {
                    const done = index < completedCriteria;
                    return (
                      <li key={criterion} className="flex items-start gap-3 rounded-2xl border bg-muted/30 p-4 text-sm leading-6">
                        {done ? <CheckSquare className="mt-0.5 shrink-0 text-primary" /> : <Square className="mt-0.5 shrink-0 text-muted-foreground" />}
                        <span>{criterion}</span>
                      </li>
                    );
                  })}
                </ul>
              )}
            </CardContent>
          </Card>
        </div>

        <Card id="sandbox" className="overflow-hidden border-slate-800 bg-slate-950 text-slate-100 xl:sticky xl:top-20 xl:h-fit">
          <CardHeader className="border-b border-slate-800">
            <CardTitle className="flex items-center gap-2 text-slate-50">
              <Terminal className="text-primary" />
              任务沙盒
            </CardTitle>
            <CardDescription className="text-slate-400">运行按钮在 Dashboard 中执行；此处展示任务代码、验收语境和静态导师提示。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5 p-6">
            <CodePreview code={mission.starterCode} locked={mission.status === "locked"} />
            <Alert className="border-cyan-900 bg-cyan-950 text-cyan-100">
              <GraduationCap />
              <AlertTitle>静态导师提示</AlertTitle>
              <AlertDescription>
                这里不会调用 AI 服务。请先阅读任务目标，再到 Dashboard 沙盒运行当前任务并根据 stdout/stderr 迭代。
              </AlertDescription>
            </Alert>
            {mission.status === "locked" ? (
              <Button className="w-full" disabled>
                任务未解锁
              </Button>
            ) : (
              <Button asChild className="w-full">
                <Link to={`/dashboard?mission=${mission.slug}#sandbox`}>打开运行工作台</Link>
              </Button>
            )}
          </CardContent>
        </Card>
      </section>
    </main>
  );
}

function MissionMeta({ icon, text }: { icon: ReactNode; text: string }) {
  return (
    <span className="inline-flex items-center gap-2">
      <span className="text-primary">{icon}</span>
      {text}
    </span>
  );
}

function MissionNavButton({ mission, label, direction }: { mission?: Mission; label: string; direction: "previous" | "next" }) {
  const content = (
    <>
      {direction === "previous" && <ArrowLeft data-icon="inline-start" />}
      {label}
      {direction === "next" && <ArrowRight data-icon="inline-end" />}
    </>
  );

  if (!mission) {
    return <Button disabled variant="outline" size="sm">{content}</Button>;
  }

  return (
    <Button asChild variant="outline" size="sm">
      <Link to={`/missions/${mission.slug}`}>{content}</Link>
    </Button>
  );
}

function MissionCta({ mission }: { mission: Mission }) {
  if (mission.status === "locked") {
    return (
      <Button disabled className="w-full">
        <Lock data-icon="inline-start" />
        完成前置任务解锁
      </Button>
    );
  }

  const label = mission.status === "completed" ? "查看复盘" : mission.status === "in-progress" ? "继续挑战" : "开始挑战";

  return (
    <Button asChild className="w-full">
      <Link to={`/dashboard?mission=${mission.slug}#sandbox`}>{label}</Link>
    </Button>
  );
}

function ListBlock({ items, icon }: { items: string[]; icon: ReactNode }) {
  return (
    <ul className="space-y-3">
      {items.map((item) => (
        <li key={item} className="flex gap-3 rounded-2xl border bg-muted/30 p-4 text-sm leading-6">
          <span className="mt-0.5 shrink-0 text-primary">{icon}</span>
          <span>{item}</span>
        </li>
      ))}
    </ul>
  );
}

function CodePreview({ code, locked }: { code: string; locked: boolean }) {
  const lines = code.split("\n");

  return (
    <div className="relative overflow-hidden rounded-2xl border border-slate-800 bg-black">
      {locked && <div className="absolute inset-0 z-10 flex items-center justify-center bg-slate-950/70 text-sm font-semibold text-slate-300 backdrop-blur-[2px]">完成前置任务后解锁代码</div>}
      <div className={locked ? "overflow-x-auto blur-sm" : "overflow-x-auto"}>
        <div className="grid min-w-[720px] grid-cols-[auto_1fr] text-sm font-mono leading-relaxed">
          <div className="select-none border-r border-slate-800 bg-slate-950 px-4 py-6 text-right text-slate-600">
            {lines.map((_, index) => <div key={index + 1}>{index + 1}</div>)}
          </div>
          <pre className="p-6 text-slate-200">
            {lines.map((line, index) => <div key={`${index}-${line}`}>{renderGoLine(line)}</div>)}
          </pre>
        </div>
      </div>
    </div>
  );
}

function renderGoLine(line: string) {
  const commentStart = line.indexOf("//");
  const code = commentStart >= 0 ? line.slice(0, commentStart) : line;
  const comment = commentStart >= 0 ? line.slice(commentStart) : "";

  return (
    <>
      {renderGoTokens(code)}
      {comment && <span className="text-slate-500 italic">{comment}</span>}
    </>
  );
}

function renderGoTokens(code: string) {
  return code.split(/("(?:\\.|[^"\\])*"|`[^`]*`|\b\d+\b|\b[A-Za-z_][A-Za-z0-9_]*\b|\s+|.)/g).filter(Boolean).map((token, index) => {
    if (/^\s+$/.test(token)) return token;
    if (/^"/.test(token) || /^`/.test(token)) return <span key={`${token}-${index}`} className="text-emerald-400">{token}</span>;
    if (/^\d+$/.test(token)) return <span key={`${token}-${index}`} className="text-violet-400">{token}</span>;
    if (goKeywords.has(token)) return <span key={`${token}-${index}`} className="text-pink-400">{token}</span>;
    if (goFunctions.has(token)) return <span key={`${token}-${index}`} className="text-cyan-300">{token}</span>;
    return token;
  });
}
