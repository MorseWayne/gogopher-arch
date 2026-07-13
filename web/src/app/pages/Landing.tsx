import type { ReactNode } from 'react'
import { Link } from 'react-router'
import {
  ArrowRight,
  Boxes,
  CalendarClock,
  CheckCircle2,
  Github,
  Layers3,
  ShieldCheck,
  Sparkles,
  Terminal,
} from 'lucide-react'

import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'

const workflow = [
  {
    icon: <Layers3 />,
    title: '冻结学习上下文',
    text: 'Attempt 固定 release、Activity、Task、workspace 与规则 hash，刷新后仍从服务端恢复同一事实。',
  },
  {
    icon: <Terminal />,
    title: '执行公开动作',
    text: 'Build、Test、Vet 与 Submit 通过 versioned multi-file runner 执行，客户端不复制验收规则。',
  },
  {
    icon: <CheckCircle2 />,
    title: '记录能力证据',
    text: 'EvaluationBatch 生成 Evidence，再由 projection 更新 Capability Snapshot；页面只展示服务端结果。',
  },
  {
    icon: <CalendarClock />,
    title: '安排后续练习',
    text: 'Learning queue 根据 prerequisite、Snapshot 与 review due time 返回下一项，而不是静态推荐。',
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
                Capability · Evidence · Review
              </Badge>
              <div className="space-y-4">
                <h1 className="max-w-4xl text-4xl font-bold tracking-tight md:text-6xl lg:text-7xl">
                  用可复现证据推进 Go 能力
                </h1>
                <p className="max-w-2xl text-lg leading-8 text-muted-foreground md:text-xl">
                  GoGopher Arch 把学习定义、代码执行、评估证据和间隔 review 连成一个服务端闭环。
                  学习进度来自可追溯 Evidence，不来自浏览器里的静态百分比。
                </p>
              </div>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row">
              <Button asChild size="lg">
                <Link to="/dashboard">
                  打开学习工作台
                  <ArrowRight />
                </Link>
              </Button>
              <Button asChild variant="outline" size="lg">
                <a href="https://github.com/MorseWayne/gogopher-arch" target="_blank" rel="noreferrer">
                  <Github />
                  查看实现
                </a>
              </Button>
            </div>

            <div className="grid gap-3 sm:grid-cols-3">
              <Metric value="Versioned" label="immutable release" />
              <Metric value="Server-side" label="Evidence projection" />
              <Metric value="Due-aware" label="review scheduling" />
            </div>
          </div>

          <Card className="border-primary/20 bg-background/90 shadow-2xl shadow-primary/10 backdrop-blur">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Boxes className="text-primary" />
                一次可追溯练习
              </CardTitle>
              <CardDescription>每一步都有明确的事实来源和失败语义。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <TraceRow index="01" title="Activity" text="公开目标、mode 与 assistance policy" />
              <TraceRow index="02" title="Attempt" text="冻结 workspace 与 release 引用" />
              <TraceRow index="03" title="Submission" text="幂等提交与基础设施 retry" />
              <TraceRow index="04" title="Evidence" text="规则结果、independence 与 context" />
              <TraceRow index="05" title="Snapshot" text="能力状态与下一次 review" />
            </CardContent>
          </Card>
        </div>
      </section>

      <section className="mx-auto w-full max-w-7xl px-4 py-16 md:px-6">
        <div className="mb-8 max-w-3xl">
          <Badge variant="outline" className="mb-3">Learning loop</Badge>
          <h2 className="text-3xl font-bold tracking-tight md:text-4xl">产品围绕证据闭环，而不是内容目录</h2>
          <p className="mt-3 leading-7 text-muted-foreground">
            内容发布、执行、评估和 review 使用同一组 versioned identifiers；前端负责交互与解释，不自行判定掌握状态。
          </p>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          {workflow.map((item) => <WorkflowCard key={item.title} {...item} />)}
        </div>
      </section>

      <section className="border-t bg-muted/25">
        <div className="mx-auto grid w-full max-w-7xl gap-6 px-4 py-14 md:px-6 lg:grid-cols-[0.8fr_1.2fr]">
          <div>
            <Badge variant="secondary" className="mb-3">
              <ShieldCheck />
              Boundary
            </Badge>
            <h2 className="text-2xl font-bold">明确安全边界</h2>
          </div>
          <div className="space-y-3 text-sm leading-7 text-muted-foreground">
            <p>匿名同源 session 只延续 Learner ownership，不等同于账号认证。</p>
            <p>held-out checks 用于减少评估内容的意外暴露，不构成防作弊边界。</p>
            <p>当前 Sandbox 面向本地可信学习环境；policy-only network 标记不代表生产级进程隔离。</p>
          </div>
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
