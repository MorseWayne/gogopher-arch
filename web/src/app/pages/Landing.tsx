import type { ReactNode } from "react";
import { Link } from "react-router";
import { ArrowRight, BookOpen, Bot, CheckCircle2, Code2, Github, Map, Rocket, ShieldCheck, Sparkles, Terminal } from "lucide-react";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Separator } from "../components/ui/separator";
import { goBasicsChapters } from "../data/goBasicsCourse";
import { missions } from "../data/missions";

const paths = [
  {
    title: "Go 基础训练营",
    description: "13 章内置课程，从第一段 Go 程序到测试、并发和工程实践。",
    href: "/courses/go-basics",
    status: "可学习",
    icon: BookOpen,
  },
  {
    title: "后端实习任务线",
    description: "围绕真实团队第一周任务，练习修 Bug、读日志、写验收。",
    href: `/missions/${missions[0].slug}`,
    status: "可挑战",
    icon: Code2,
  },
  {
    title: "工程能力进阶",
    description: "数据库、缓存、并发、可观测性和部署可靠性路线。",
    status: "即将开放",
    icon: ShieldCheck,
  },
  {
    title: "AI 全栈路线",
    description: "LLM API、RAG、Agent、评测和 AI 产品工程能力。",
    status: "即将开放",
    icon: Bot,
  },
];

export function Landing() {
  return (
    <div className="bg-background">
      <section className="relative overflow-hidden border-b">
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_left,_rgba(0,173,216,0.18),_transparent_34%)]" />
        <div className="mx-auto grid min-h-[calc(100svh-4rem)] w-full max-w-7xl items-center gap-12 px-4 py-20 md:px-6 lg:grid-cols-[minmax(0,1.08fr)_minmax(360px,0.92fr)]">
          <div className="flex flex-col gap-8">
            <div className="flex flex-col gap-5">
              <Badge variant="secondary" className="w-fit">
                <Sparkles data-icon="inline-start" />
                Go 后端实习成长平台
              </Badge>
              <div className="flex flex-col gap-4">
                <h1 className="max-w-4xl text-4xl font-bold tracking-tight text-foreground md:text-6xl lg:text-7xl">
                  从 Go 基础到真实后端任务
                </h1>
                <p className="max-w-2xl text-lg leading-8 text-muted-foreground md:text-xl">
                  GoGopher Arch 把课程正文、浏览器沙盒和虚拟职场任务放在同一条路径里，帮助你从第一行 Go 代码走到后端实习基本功。
                </p>
              </div>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row">
              <Button asChild size="lg">
                <Link to="/courses/go-basics">
                  开始 Go 基础训练营
                  <ArrowRight data-icon="inline-end" />
                </Link>
              </Button>
              <Button asChild variant="outline" size="lg">
                <Link to="/dashboard">进入学习总览</Link>
              </Button>
            </div>

            <div className="grid gap-3 sm:grid-cols-3">
              <HeroMetric label="内置章节" value={`${goBasicsChapters.length}`} />
              <HeroMetric label="实习任务" value={`${missions.length}`} />
              <HeroMetric label="沙盒链路" value="可运行" />
            </div>
          </div>

          <Card className="border-primary/20 bg-background/90 shadow-2xl shadow-primary/10 backdrop-blur">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Terminal className="text-primary" />
                今日推荐路径
              </CardTitle>
              <CardDescription>先补齐 Go 基础，再进入任务线动手修复真实问题。</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-5">
              <div className="rounded-2xl border bg-muted/40 p-4">
                <div className="mb-3 flex items-center justify-between gap-3">
                  <div>
                    <div className="font-semibold">Go 基础训练营 · 第 1 章</div>
                    <div className="text-sm text-muted-foreground">从可运行程序入口开始</div>
                  </div>
                  <Badge>推荐</Badge>
                </div>
                <div className="h-2 rounded-full bg-muted">
                  <div className="h-2 w-[18%] rounded-full bg-primary" />
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <MiniStep icon={<BookOpen />} title="阅读正文" text="学习目标、现代 Go 说明、工程实践" />
                <MiniStep icon={<Terminal />} title="运行练习" text="在沙盒中查看 stdout/stderr/exit code" />
                <MiniStep icon={<Code2 />} title="挑战任务" text="进入实习任务线修复真实问题" />
                <MiniStep icon={<Map />} title="查看路线" text="工程进阶与 AI 全栈路线标记为即将开放" />
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      <section id="paths" className="mx-auto flex w-full max-w-7xl flex-col gap-8 px-4 py-16 md:px-6">
        <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
          <div>
            <Badge variant="outline" className="mb-3">学习路径</Badge>
            <h2 className="text-3xl font-bold tracking-tight md:text-4xl">一条清晰的成长路线</h2>
            <p className="mt-3 max-w-2xl text-muted-foreground">已实现的入口可以直接学习或挑战；未实现路线只做预告，不制造假功能。</p>
          </div>
          <Button asChild variant="outline">
            <a href="https://github.com/MorseWayne/gogopher-arch" target="_blank" rel="noreferrer">
              <Github data-icon="inline-start" />
              查看源码
            </a>
          </Button>
        </div>

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {paths.map((path) => (
            <PathCard key={path.title} path={path} />
          ))}
        </div>
      </section>

      <Separator />

      <section className="mx-auto grid w-full max-w-7xl gap-6 px-4 py-16 md:px-6 lg:grid-cols-[0.9fr_1.1fr]">
        <div className="flex flex-col gap-4">
          <Badge variant="secondary" className="w-fit">
            <Rocket data-icon="inline-start" />
            MVP 闭环
          </Badge>
          <h2 className="text-3xl font-bold tracking-tight">不是单纯看文档，而是边学边运行</h2>
          <p className="text-muted-foreground leading-7">
            课程页采用清爽阅读布局，任务页保留工作台和终端反馈。你会看到每段代码的输出、错误和下一步建议，而不是只在目录里跳转。
          </p>
        </div>
        <div className="grid gap-4 sm:grid-cols-3">
          <Feature icon={<CheckCircle2 />} title="内置课程" text="内容完整留在应用内，外部资料只作参考。" />
          <Feature icon={<Terminal />} title="沙盒反馈" text="运行结果、错误和 timeout 是学习反馈。" />
          <Feature icon={<ShieldCheck />} title="无假功能" text="Coming soon 路线明确标识，不伪装成已实现能力。" />
        </div>
      </section>
    </div>
  );
}

function HeroMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border bg-background/70 p-4 shadow-sm">
      <div className="text-2xl font-bold text-foreground">{value}</div>
      <div className="text-sm text-muted-foreground">{label}</div>
    </div>
  );
}

function MiniStep({ icon, title, text }: { icon: ReactNode; title: string; text: string }) {
  return (
    <div className="rounded-xl border bg-background p-3">
      <div className="mb-2 flex items-center gap-2 font-medium">
        <span className="text-primary">{icon}</span>
        {title}
      </div>
      <p className="text-sm text-muted-foreground">{text}</p>
    </div>
  );
}

function PathCard({ path }: { path: (typeof paths)[number] }) {
  const Icon = path.icon;
  const available = Boolean(path.href);
  const card = (
    <Card className="h-full transition-colors hover:border-primary/40">
      <CardHeader>
        <div className="mb-4 flex items-center justify-between gap-3">
          <span className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Icon />
          </span>
          <Badge variant={available ? "default" : "secondary"}>{path.status}</Badge>
        </div>
        <CardTitle>{path.title}</CardTitle>
        <CardDescription>{path.description}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="inline-flex items-center gap-2 text-sm font-medium text-primary">
          {available ? "进入路径" : "即将开放"}
          {available && <ArrowRight className="size-4" />}
        </div>
      </CardContent>
    </Card>
  );

  if (!path.href) {
    return <div aria-disabled="true">{card}</div>;
  }

  return <Link to={path.href}>{card}</Link>;
}

function Feature({ icon, title, text }: { icon: ReactNode; title: string; text: string }) {
  return (
    <Card>
      <CardHeader>
        <div className="mb-2 text-primary">{icon}</div>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm leading-6 text-muted-foreground">{text}</p>
      </CardContent>
    </Card>
  );
}
