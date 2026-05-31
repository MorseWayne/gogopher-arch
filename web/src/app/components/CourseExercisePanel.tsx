import { lazy, Suspense, useEffect, useMemo, useState } from "react";
import { CheckCircle2, Clock3, Code2, Lightbulb, Loader2, Play, RotateCcw, Terminal, XCircle } from "lucide-react";
import { executeCode } from "../../api/execute";
import type { SandboxResponse } from "../../api/types";
import type { GoCourseExercise } from "../data/goBasicsCourse";
import { Alert, AlertDescription, AlertTitle } from "./ui/alert";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";
import { cn } from "./ui/utils";

const GoCodeEditor = lazy(() => import("./GoCodeEditor").then((module) => ({ default: module.GoCodeEditor })));

export function CourseExercisePanel({ chapterSlug, exercise, exercises }: { chapterSlug: string; exercise: GoCourseExercise; exercises?: GoCourseExercise[] }) {
  const exerciseList = useMemo(() => normalizeExercises(exercise, exercises), [exercise, exercises]);
  const [activeExerciseId, setActiveExerciseId] = useState(getExerciseId(exerciseList[0], 0));
  const activeIndex = Math.max(
    0,
    exerciseList.findIndex((item, index) => getExerciseId(item, index) === activeExerciseId),
  );
  const activeExercise = exerciseList[activeIndex] ?? exerciseList[0];
  const activeKey = getExerciseId(activeExercise, activeIndex);

  const [codeByExerciseId, setCodeByExerciseId] = useState<Record<string, string>>({});
  const [isRunning, setIsRunning] = useState(false);
  const [result, setResult] = useState<SandboxResponse | null>(null);
  const [error, setError] = useState("");
  const [visibleHintCount, setVisibleHintCount] = useState(1);

  const currentCode = codeByExerciseId[activeKey] ?? activeExercise.starterCode;
  const isOutputMatched = result ? matchesExpectedOutput(result.stdout, activeExercise.expectedOutput, activeExercise.outputMatch) : false;
  const isExecutionSuccessful = result?.exit_code === 0;
  const storageKey = `go-basics:v2:${chapterSlug}:${activeKey}`;
  const visibleHints = activeExercise.hints.slice(0, visibleHintCount);
  const hasMoreHints = visibleHintCount < activeExercise.hints.length;

  useEffect(() => {
    const nextId = getExerciseId(exerciseList[0], 0);
    setActiveExerciseId((currentId) => (exerciseList.some((item, index) => getExerciseId(item, index) === currentId) ? currentId : nextId));
  }, [exerciseList]);

  useEffect(() => {
    if (typeof window === "undefined") {
      setCodeByExerciseId((current) => ({ ...current, [activeKey]: activeExercise.starterCode }));
      return;
    }

    const savedCode = window.localStorage.getItem(storageKey);
    setCodeByExerciseId((current) => ({
      ...current,
      [activeKey]: savedCode ?? current[activeKey] ?? activeExercise.starterCode,
    }));
  }, [activeExercise.starterCode, activeKey, storageKey]);

  useEffect(() => {
    setResult(null);
    setError("");
    setVisibleHintCount(1);
  }, [activeKey]);

  const handleCodeChange = (nextCode: string) => {
    setCodeByExerciseId((current) => ({ ...current, [activeKey]: nextCode }));

    if (typeof window !== "undefined") {
      window.localStorage.setItem(storageKey, nextCode);
    }
  };

  const handleReset = () => {
    setCodeByExerciseId((current) => ({ ...current, [activeKey]: activeExercise.starterCode }));
    setResult(null);
    setError("");
    setVisibleHintCount(1);

    if (typeof window !== "undefined") {
      window.localStorage.removeItem(storageKey);
    }
  };

  const handleRun = async () => {
    setIsRunning(true);
    setResult(null);
    setError("");

    try {
      const response = await executeCode({
        id: `${chapterSlug}-${activeKey}-${Date.now()}`,
        code: currentCode,
        language: "go",
        timeout: 3,
      });
      setResult(response);
    } catch (runError) {
      setError(runError instanceof Error ? runError.message : "无法连接到代码运行服务");
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <Card className="overflow-hidden border-slate-800 bg-slate-950 text-slate-100 shadow-xl shadow-primary/10">
      <CardHeader className="border-b border-slate-800">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <CardTitle className="flex items-center gap-2 text-slate-50">
              <Terminal className="text-primary" />
              沙盒练习
            </CardTitle>
            <CardDescription className="mt-2 text-slate-400">选择练习、编辑代码并运行。编辑器支持 Go 高亮和课程提示，草稿会保存在当前浏览器中。</CardDescription>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button onClick={handleReset} disabled={isRunning} variant="outline" className="border-slate-700 bg-slate-900 text-slate-100 hover:bg-slate-800 hover:text-slate-50">
              <RotateCcw data-icon="inline-start" />
              重置
            </Button>
            <Button onClick={handleRun} disabled={isRunning}>
              {isRunning ? <Loader2 className="animate-spin" data-icon="inline-start" /> : <Play data-icon="inline-start" />}
              {isRunning ? "运行中" : "运行代码"}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-5 p-6">
        {exerciseList.length > 1 && (
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
            {exerciseList.map((item, index) => {
              const itemId = getExerciseId(item, index);
              const isActive = itemId === activeKey;

              return (
                <button
                  key={itemId}
                  type="button"
                  onClick={() => setActiveExerciseId(itemId)}
                  className={cn(
                    "rounded-2xl border p-4 text-left transition hover:border-primary/80 hover:bg-slate-900",
                    isActive ? "border-primary bg-primary/10" : "border-slate-800 bg-slate-900/70",
                  )}
                >
                  <div className="mb-2 flex flex-wrap gap-2">
                    <Badge variant={isActive ? "default" : "secondary"}>{formatExerciseDifficulty(item.difficulty)}</Badge>
                    <Badge variant="outline" className="border-slate-700 text-slate-300">{formatExerciseKind(item.kind)}</Badge>
                  </div>
                  <div className="font-semibold text-slate-50">{item.title}</div>
                  <div className="mt-2 line-clamp-2 text-xs leading-5 text-slate-400">{item.prompt}</div>
                </button>
              );
            })}
          </div>
        )}

        <section className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
          <div className="mb-3 flex flex-wrap items-start justify-between gap-3">
            <div>
              <div className="flex flex-wrap gap-2">
                <Badge>{formatExerciseDifficulty(activeExercise.difficulty)}</Badge>
                <Badge variant="secondary">{formatExerciseKind(activeExercise.kind)}</Badge>
                {activeExercise.estimatedMinutes && (
                  <Badge variant="outline" className="border-slate-700 text-slate-300">
                    <Clock3 data-icon="inline-start" />
                    {activeExercise.estimatedMinutes} 分钟
                  </Badge>
                )}
              </div>
              <h3 className="mt-3 text-xl font-semibold text-slate-50">{activeExercise.title}</h3>
              <p className="mt-2 text-sm leading-6 text-slate-400">{activeExercise.prompt}</p>
            </div>
          </div>

          {activeExercise.context && <p className="rounded-xl border border-slate-800 bg-black/30 p-3 text-sm leading-6 text-slate-300">{activeExercise.context}</p>}

          {activeExercise.concepts && activeExercise.concepts.length > 0 && (
            <div className="mt-4 flex flex-wrap gap-2">
              {activeExercise.concepts.map((concept) => (
                <Badge key={concept} variant="outline" className="border-slate-700 text-slate-300">{concept}</Badge>
              ))}
            </div>
          )}
        </section>

        <div>
          <div className="mb-2 flex items-center justify-between text-xs text-slate-500">
            <span>CodeMirror editor</span>
            <span>Go · 高亮/补全 · local draft</span>
          </div>
          <Suspense fallback={<div className="min-h-[28rem] rounded-2xl border border-slate-800 bg-black p-4 font-mono text-sm text-slate-500">正在加载 CodeMirror 编辑器…</div>}>
            <GoCodeEditor value={currentCode} onChange={handleCodeChange} concepts={activeExercise.concepts} />
          </Suspense>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-4">
            <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">期望输出</div>
            <code className="whitespace-pre-wrap text-sm text-cyan-300">{activeExercise.expectedOutput}</code>
          </div>
          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-4">
            <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">匹配规则</div>
            <Badge variant="secondary">{activeExercise.outputMatch}</Badge>
          </div>
        </div>

        {activeExercise.hints.length > 0 && (
          <Alert className="border-cyan-900 bg-cyan-950 text-cyan-100">
            <Lightbulb />
            <AlertTitle>渐进提示</AlertTitle>
            <AlertDescription>
              <ul className="mt-2 list-disc pl-5">
                {visibleHints.map((hint) => (
                  <li key={hint}>{hint}</li>
                ))}
              </ul>
              {hasMoreHints && (
                <Button size="sm" variant="outline" className="mt-3 border-cyan-800 bg-cyan-950 text-cyan-100 hover:bg-cyan-900" onClick={() => setVisibleHintCount((count) => count + 1)}>
                  显示下一个提示
                </Button>
              )}
            </AlertDescription>
          </Alert>
        )}

        {error && (
          <Alert variant="destructive">
            <XCircle />
            <AlertTitle>无法连接到代码运行服务</AlertTitle>
            <AlertDescription>{error}。请确认本地 Gateway 和 Sandbox Engine 已启动。</AlertDescription>
          </Alert>
        )}

        {result && (
          <div className="space-y-3 rounded-2xl border border-slate-800 bg-black p-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div className="text-sm font-semibold text-slate-50">运行结果</div>
              <div className="flex flex-wrap gap-2">
                <Badge variant={isExecutionSuccessful ? "default" : "destructive"}>exit {result.exit_code}</Badge>
                <Badge variant="secondary">{result.status}</Badge>
                <Badge variant="secondary">{result.duration}ms</Badge>
              </div>
            </div>

            <OutputBlock label="stdout" value={result.stdout} />
            <OutputBlock label="stderr" value={result.stderr} />

            <Alert className={isExecutionSuccessful && isOutputMatched ? "border-emerald-900 bg-emerald-950 text-emerald-100" : "border-amber-900 bg-amber-950 text-amber-100"}>
              {isExecutionSuccessful && isOutputMatched ? <CheckCircle2 /> : <XCircle />}
              <AlertTitle>{isExecutionSuccessful && isOutputMatched ? "练习输出已匹配" : "请检查运行状态或输出内容"}</AlertTitle>
              <AlertDescription>编译失败、输出不匹配或 timeout 都属于正常学习反馈，草稿不会被清空。</AlertDescription>
            </Alert>
          </div>
        )}

        {activeExercise.solutionOutline && activeExercise.solutionOutline.length > 0 && (
          <section className="rounded-2xl border border-slate-800 bg-slate-900 p-4">
            <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-50">
              <Code2 className="text-primary" />
              解题思路检查点
            </div>
            <ul className="list-disc space-y-2 pl-5 text-sm leading-6 text-slate-400">
              {activeExercise.solutionOutline.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </section>
        )}
      </CardContent>
    </Card>
  );
}

function OutputBlock({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="mb-1 text-xs font-semibold uppercase tracking-wider text-slate-500">{label}</div>
      <pre className="min-h-12 whitespace-pre-wrap rounded-xl border border-slate-800 bg-slate-950 p-3 text-sm text-slate-100">
        {value || <span className="text-slate-600">无输出</span>}
      </pre>
    </div>
  );
}

function normalizeExercises(exercise: GoCourseExercise, exercises?: GoCourseExercise[]) {
  return exercises && exercises.length > 0 ? exercises : [exercise];
}

function getExerciseId(exercise: GoCourseExercise, index: number) {
  return exercise.id ?? `exercise-${index + 1}`;
}

function formatExerciseKind(kind: GoCourseExercise["kind"] = "run") {
  const labels: Record<NonNullable<GoCourseExercise["kind"]>, string> = {
    run: "运行",
    edit: "改写",
    test: "测试",
    debug: "Debug",
    project: "项目",
    review: "评审",
  };

  return labels[kind];
}

function formatExerciseDifficulty(difficulty: GoCourseExercise["difficulty"] = "warmup") {
  const labels: Record<NonNullable<GoCourseExercise["difficulty"]>, string> = {
    warmup: "热身",
    core: "核心",
    challenge: "挑战",
  };

  return labels[difficulty];
}

function matchesExpectedOutput(stdout: string, expectedOutput: string, outputMatch: GoCourseExercise["outputMatch"]) {
  if (outputMatch === "trimmed-exact") {
    return stdout.trim() === expectedOutput.trim();
  }

  return stdout.includes(expectedOutput);
}
