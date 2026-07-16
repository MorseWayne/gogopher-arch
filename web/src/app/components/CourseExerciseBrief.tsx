import { ArrowRight, CheckCircle2, Code2, Lightbulb } from "lucide-react";
import { Link } from "react-router";

import type { GoCourseExercise } from "../data/goBasicsCourse";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";

export function CourseExerciseBrief({ exercise, exercises }: { exercise: GoCourseExercise; exercises?: GoCourseExercise[] }) {
  const exerciseList = exercises && exercises.length > 0 ? exercises : [exercise];

  return (
    <Card className="overflow-hidden border-slate-800 bg-slate-950 text-slate-100 shadow-xl shadow-primary/10">
      <CardHeader className="border-b border-slate-800">
        <CardTitle className="flex items-center gap-2 text-slate-50">
          <Code2 className="text-primary" />
          章节练习
        </CardTitle>
        <CardDescription className="text-slate-400">
          先阅读题目、起始代码和验收目标，再在本地 Go 环境完成。需要平台内运行、保存和反馈时，请进入学习工作台。
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5 p-6">
        {exerciseList.map((item, index) => (
          <section key={item.id ?? `${item.title}-${index}`} className="space-y-4 rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <div className="mb-2 flex flex-wrap gap-2">
                  <Badge>{formatExerciseDifficulty(item.difficulty)}</Badge>
                  <Badge variant="secondary">{formatExerciseKind(item.kind)}</Badge>
                  {item.estimatedMinutes && <Badge variant="outline" className="border-slate-700 text-slate-300">约 {item.estimatedMinutes} 分钟</Badge>}
                </div>
                <h3 className="text-xl font-semibold text-slate-50">{item.title}</h3>
                <p className="mt-2 text-sm leading-6 text-slate-400">{item.prompt}</p>
              </div>
            </div>

            {item.context && <p className="rounded-xl border border-slate-800 bg-black/30 p-3 text-sm leading-6 text-slate-300">{item.context}</p>}

            {item.concepts && item.concepts.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {item.concepts.map((concept) => <Badge key={concept} variant="outline" className="border-slate-700 text-slate-300">{concept}</Badge>)}
              </div>
            )}

            <div>
              <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">起始代码</div>
              <pre className="overflow-auto rounded-xl border border-slate-800 bg-black p-4 text-sm leading-6 text-slate-100"><code>{item.starterCode}</code></pre>
            </div>

            <div className="grid gap-3 md:grid-cols-2">
              <div className="rounded-xl border border-slate-800 bg-black/30 p-4">
                <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">期望输出</div>
                <code className="whitespace-pre-wrap text-sm text-cyan-300">{item.expectedOutput}</code>
              </div>
              <div className="rounded-xl border border-slate-800 bg-black/30 p-4">
                <div className="mb-2 flex items-center gap-2 text-sm font-semibold text-slate-100"><Lightbulb className="size-4 text-cyan-300" />提示</div>
                <ul className="list-disc space-y-1 pl-5 text-sm leading-6 text-slate-400">
                  {item.hints.map((hint) => <li key={hint}>{hint}</li>)}
                </ul>
              </div>
            </div>

            {item.solutionOutline && item.solutionOutline.length > 0 && (
              <details className="rounded-xl border border-slate-800 bg-black/20 p-4">
                <summary className="cursor-pointer text-sm font-semibold text-slate-100">查看解题思路检查点</summary>
                <ul className="mt-3 list-disc space-y-2 pl-5 text-sm leading-6 text-slate-400">
                  {item.solutionOutline.map((step) => <li key={step}>{step}</li>)}
                </ul>
              </details>
            )}
          </section>
        ))}

        <div className="flex flex-col gap-3 rounded-2xl border border-emerald-900 bg-emerald-950/60 p-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex gap-3 text-sm leading-6 text-emerald-100">
            <CheckCircle2 className="mt-0.5 shrink-0 text-emerald-300" />
            <span>课程目录用于系统查阅；工作台提供可执行任务、自动验收和学习进度。</span>
          </div>
          <Button asChild className="shrink-0">
            <Link to="/dashboard">进入学习工作台<ArrowRight /></Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function formatExerciseKind(kind: GoCourseExercise["kind"] = "run") {
  return ({ run: "运行", edit: "改写", test: "测试", debug: "Debug", project: "项目", review: "评审" } as const)[kind];
}

function formatExerciseDifficulty(difficulty: GoCourseExercise["difficulty"] = "warmup") {
  return ({ warmup: "热身", core: "核心", challenge: "挑战" } as const)[difficulty];
}
