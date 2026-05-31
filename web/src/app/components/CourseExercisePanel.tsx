import { useState } from "react";
import { CheckCircle2, Loader2, Play, Terminal, XCircle } from "lucide-react";
import { executeCode } from "../../api/execute";
import type { SandboxResponse } from "../../api/types";
import type { GoCourseExercise } from "../data/goBasicsCourse";
import { Alert, AlertDescription, AlertTitle } from "./ui/alert";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";

export function CourseExercisePanel({ chapterSlug, exercise }: { chapterSlug: string; exercise: GoCourseExercise }) {
  const [isRunning, setIsRunning] = useState(false);
  const [result, setResult] = useState<SandboxResponse | null>(null);
  const [error, setError] = useState("");

  const isOutputMatched = result ? matchesExpectedOutput(result.stdout, exercise.expectedOutput, exercise.outputMatch) : false;
  const isExecutionSuccessful = result?.exit_code === 0;

  const handleRun = async () => {
    setIsRunning(true);
    setResult(null);
    setError("");

    try {
      const response = await executeCode({
        id: `${chapterSlug}-${Date.now()}`,
        code: exercise.starterCode,
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
              {exercise.title}
            </CardTitle>
            <CardDescription className="mt-2 text-slate-400">{exercise.prompt}</CardDescription>
          </div>
          <Button onClick={handleRun} disabled={isRunning}>
            {isRunning ? <Loader2 className="animate-spin" data-icon="inline-start" /> : <Play data-icon="inline-start" />}
            {isRunning ? "运行中" : "运行代码"}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-5 p-6">
        <div>
          <div className="mb-2 flex items-center justify-between text-xs text-slate-500">
            <span>starter code</span>
            <span>Go · timeout 3s</span>
          </div>
          <pre className="max-h-80 overflow-auto rounded-2xl border border-slate-800 bg-black p-4 text-sm leading-6 text-slate-100">
            <code>{exercise.starterCode}</code>
          </pre>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-4">
            <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">期望输出</div>
            <code className="text-sm text-cyan-300">{exercise.expectedOutput}</code>
          </div>
          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-4">
            <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">匹配规则</div>
            <Badge variant="secondary">{exercise.outputMatch}</Badge>
          </div>
        </div>

        {exercise.hints.length > 0 && (
          <Alert className="border-cyan-900 bg-cyan-950 text-cyan-100">
            <Terminal />
            <AlertTitle>提示</AlertTitle>
            <AlertDescription>
              <ul className="mt-2 list-disc pl-5">
                {exercise.hints.map((hint) => (
                  <li key={hint}>{hint}</li>
                ))}
              </ul>
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
              <AlertDescription>编译失败、输出不匹配或 timeout 都属于正常学习反馈，不会清空你的代码。</AlertDescription>
            </Alert>
          </div>
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

function matchesExpectedOutput(stdout: string, expectedOutput: string, outputMatch: GoCourseExercise["outputMatch"]) {
  if (outputMatch === "trimmed-exact") {
    return stdout.trim() === expectedOutput.trim();
  }

  return stdout.includes(expectedOutput);
}
