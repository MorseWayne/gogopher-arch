import type { ReactNode } from 'react'
import { Link } from 'react-router'
import {
  ArrowRight,
  BookOpenText,
  CheckCircle2,
  Github,
  Route,
  Sparkles,
  Terminal,
} from 'lucide-react'

import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'

const workflow = [
  {
    icon: <BookOpenText />,
    title: '先理解一个具体问题',
    text: '每节课先解释为什么要学、工具在解决什么问题，以及完成后你应该能够做什么。',
  },
  {
    icon: <Terminal />,
    title: '马上在真实代码里练习',
    text: '在浏览器工作区运行 Build、Test、Vet，直接根据工具反馈调整判断和代码。',
  },
  {
    icon: <CheckCircle2 />,
    title: '得到可以理解的反馈',
    text: '你会看到哪些检查通过、哪里仍需改进，而不是只得到一个模糊的完成百分比。',
  },
  {
    icon: <Route />,
    title: '继续最合适的下一步',
    text: '进行中的练习可以随时恢复；完成后，学习台会根据结果安排下一节或复习。',
  },
]

export function Landing() {
  return (
    <div className="bg-background">
      <section className="relative overflow-hidden border-b">
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_left,_color-mix(in_oklab,var(--color-primary)_22%,transparent),_transparent_38%)]" />
        <div className="mx-auto grid min-h-[calc(100svh-4rem)] w-full max-w-7xl items-center gap-12 px-4 py-20 md:px-6 lg:grid-cols-[minmax(0,1.08fr)_minmax(360px,0.92fr)]">
          <div className="space-y-8">
            <div className="space-y-5">
              <Badge variant="secondary" className="w-fit">
                <Sparkles />
                从 Go 基础走向独立开发
              </Badge>
              <div className="space-y-4">
                <h1 className="max-w-4xl text-4xl font-bold tracking-tight md:text-6xl lg:text-7xl">
                  从会写 Go 语法，到能独立完成程序
                </h1>
                <p className="max-w-2xl text-lg leading-8 text-muted-foreground md:text-xl">
                  面向已经接触过 Go 基础、但还不熟悉完整工程实践的开发者。
                  通过短讲解、真实代码练习和明确反馈，一步步建立独立解决问题的能力。
                </p>
              </div>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row">
              <Button asChild size="lg">
                <Link to="/dashboard">
                  开始第一节
                  <ArrowRight />
                </Link>
              </Button>
              <Button asChild variant="outline" size="lg">
                <a href="https://github.com/MorseWayne/gogopher-arch" target="_blank" rel="noreferrer">
                  <Github />
                  查看开源项目
                </a>
              </Button>
            </div>

            <div className="grid gap-3 sm:grid-cols-3">
              <Metric value="先理解" label="知道为什么这样做" />
              <Metric value="再动手" label="在真实代码中验证" />
              <Metric value="有反馈" label="清楚下一步改什么" />
            </div>
          </div>

          <Card className="border-primary/20 bg-background/90 shadow-2xl shadow-primary/10 backdrop-blur">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Route className="text-primary" />
                第一节会怎样进行
              </CardTitle>
              <CardDescription>从理解工具反馈开始，完成一条最小学习闭环。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <TraceRow index="01" title="阅读" text="理解 Build、Test、Vet 的职责" />
              <TraceRow index="02" title="运行" text="亲手运行三个工具并阅读输出" />
              <TraceRow index="03" title="总结" text="用自己的话说明区别和排查顺序" />
              <TraceRow index="04" title="反馈" text="查看最终检查结果与掌握情况" />
              <TraceRow index="05" title="继续" text="进入下一项适合你的练习" />
            </CardContent>
          </Card>
        </div>
      </section>

      <section className="mx-auto w-full max-w-7xl px-4 py-16 md:px-6">
        <div className="mb-8 max-w-3xl">
          <Badge variant="outline" className="mb-3">学习方式</Badge>
          <h2 className="text-3xl font-bold tracking-tight md:text-4xl">不是“看完了”，而是真的做到了</h2>
          <p className="mt-3 leading-7 text-muted-foreground">
            阅读只是开始。每个目标都会落到一次可运行、可测试、可解释的练习中。
          </p>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          {workflow.map((item) => <WorkflowCard key={item.title} {...item} />)}
        </div>
      </section>

      <section className="border-t bg-muted/25">
        <div className="mx-auto flex w-full max-w-7xl flex-col items-start justify-between gap-5 px-4 py-14 md:flex-row md:items-center md:px-6">
          <div>
            <Badge variant="secondary" className="mb-3">现在开始</Badge>
            <h2 className="text-2xl font-bold">先完成一节，再决定下一步</h2>
            <p className="mt-2 text-sm leading-7 text-muted-foreground">不需要先规划整条路线，学习台会保存进度并带你继续。</p>
          </div>
          <Button asChild size="lg"><Link to="/dashboard">进入学习台<ArrowRight /></Link></Button>
        </div>
      </section>
    </div>
  )
}

function Metric({ value, label }: { value: string; label: string }) {
  return (
    <div className="rounded-2xl border bg-background/70 p-4 shadow-sm">
      <div className="font-semibold">{value}</div>
      <div className="mt-1 text-xs text-muted-foreground">{label}</div>
    </div>
  )
}

function TraceRow({ index, title, text }: { index: string; title: string; text: string }) {
  return (
    <div className="grid grid-cols-[2rem_6rem_1fr] items-center gap-3 rounded-xl border bg-muted/25 p-3 text-sm">
      <span className="font-mono text-xs text-primary">{index}</span>
      <span className="font-semibold">{title}</span>
      <span className="text-muted-foreground">{text}</span>
    </div>
  )
}

function WorkflowCard({ icon, title, text }: { icon: ReactNode; title: string; text: string }) {
  return (
    <Card>
      <CardHeader>
        <div className="mb-2 text-primary">{icon}</div>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="leading-7 text-muted-foreground">{text}</p>
      </CardContent>
    </Card>
  )
}
