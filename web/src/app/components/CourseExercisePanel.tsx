import { useState } from "react";
import { CheckCircle2, Loader2, Play, Terminal, XCircle } from "lucide-react";
import { executeCode } from "../../api/execute";
import type { SandboxResponse } from "../../api/types";
import type { GoCourseExercise } from "../data/goBasicsCourse";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

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
    <Card className="border-neutral-800 bg-neutral-900/80 text-neutral-100 shadow-2xl shadow-[#00ADD8]/5">
      <CardHeader className="border-b border-neutral-800">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-2">
            <CardTitle className="flex items-center gap-2 text-xl text-white">
              <Terminal className="h-5 w-5 text-[#00ADD8]" />
              {exercise.title}
            </CardTitle>
            <p className="text-sm leading-6 text-neutral-400">{exercise.prompt}</p>
          </div>
          <Button onClick={handleRun} disabled={isRunning} className="rounded-xl bg-[#00ADD8] font-bold text-neutral-950 hover:bg-[#00ADD8]/90">
            {isRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
            {isRunning ? "运行中" : "运行代码"}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-5 p-6">
        <div>
          <div className="mb-2 flex items-center justify-between text-xs text-neutral-500">
            <span>starter code</span>
            <span>Go · timeout 3s</span>
          </div>
          <pre className="max-h-80 overflow-auto rounded-xl border border-neutral-800 bg-neutral-950 p-4 text-sm leading-6 text-neutral-200">
            <code>{exercise.starterCode}</code>
          </pre>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          <div className="rounded-xl border border-neutral-800 bg-neutral-950 p-4">
            <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-neutral-500">期望输出</div>
            <code className="text-sm text-[#00ADD8]">{exercise.expectedOutput}</code>
          </div>
          <div className="rounded-xl border border-neutral-800 bg-neutral-950 p-4">
            <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-neutral-500">匹配规则</div>
            <Badge className="border-neutral-700 bg-neutral-800 text-neutral-300">{exercise.outputMatch}</Badge>
          </div>
        </div>

        {exercise.hints.length > 0 && (
          <div className="rounded-xl border border-blue-500/20 bg-blue-500/5 p-4">
            <div className="mb-2 text-sm font-semibold text-blue-200">提示</div>
            <ul className="space-y-1 text-sm text-blue-100/80">
              {exercise.hints.map((hint) => (
                <li key={hint}>• {hint}</li>
              ))}
            </ul>
          </div>
        )}

        {error && (
          <div className="rounded-xl border border-red-500/30 bg-red-500/10 p-4 text-sm text-red-200">
            <div className="mb-1 flex items-center gap-2 font-semibold">
              <XCircle className="h-4 w-4" />
              无法连接到代码运行服务
            </div>
            <p className="text-red-100/80">{error}</p>
            <p className="mt-2 text-red-100/70">请确认本地 Gateway 和 Sandbox Engine 已启动。</p>
          </div>
        )}

        {result && (
          <div className="space-y-3 rounded-xl border border-neutral-800 bg-neutral-950 p-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div className="text-sm font-semibold text-white">运行结果</div>
              <div className="flex flex-wrap gap-2">
                <Badge className={isExecutionSuccessful ? "border-green-500/30 bg-green-500/10 text-green-300" : "border-red-500/30 bg-red-500/10 text-red-300"}>
                  exit {result.exit_code}
                </Badge>
                <Badge className="border-neutral-700 bg-neutral-800 text-neutral-300">{result.status}</Badge>
                <Badge className="border-neutral-700 bg-neutral-800 text-neutral-300">{result.duration}ms</Badge>
              </div>
            </div>

            <OutputBlock label="stdout" value={result.stdout} />
            <OutputBlock label="stderr" value={result.stderr} />

            <div className={isExecutionSuccessful && isOutputMatched ? "rounded-lg border border-green-500/30 bg-green-500/10 p-3 text-sm text-green-200" : "rounded-lg border border-yellow-500/30 bg-yellow-500/10 p-3 text-sm text-yellow-200"}>
              <div className="flex items-center gap-2 font-semibold">
                {isExecutionSuccessful && isOutputMatched ? <CheckCircle2 className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
                {isExecutionSuccessful && isOutputMatched ? "练习输出已匹配" : "请检查运行状态或输出内容"}
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function OutputBlock({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="mb-1 text-xs font-semibold uppercase tracking-wider text-neutral-500">{label}</div>
      <pre className="min-h-12 whitespace-pre-wrap rounded-lg border border-neutral-800 bg-black p-3 text-sm text-neutral-200">
        {value || <span className="text-neutral-600">无输出</span>}
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
