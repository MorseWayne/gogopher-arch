import { Link } from "react-router";
import { ArrowRight, BookOpen, CheckCircle2, Clock3, GraduationCap, ShieldCheck, Star } from "lucide-react";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { goBasicsChapters, validateGoBasicsCourse } from "../data/goBasicsCourse";

const validationErrors = validateGoBasicsCourse();

export function GoBasicsCourse() {
  return (
    <main className="flex-1 bg-neutral-950 text-neutral-50">
      <section className="border-b border-neutral-900 bg-[radial-gradient(circle_at_top,_rgba(0,173,216,0.16),_transparent_40%),#0a0a0a] px-6 py-20">
        <div className="container mx-auto max-w-6xl">
          <div className="grid gap-10 lg:grid-cols-[minmax(0,3fr)_minmax(320px,2fr)] lg:items-center">
            <div className="space-y-6">
              <Badge className="border-[#00ADD8]/30 bg-[#00ADD8]/10 text-[#00ADD8]">Go 基础训练营 · 13 章完整路径</Badge>
              <div className="space-y-4">
                <h1 className="text-4xl font-extrabold tracking-tight text-white md:text-6xl">从第一行 Go 代码到后端实习基本功</h1>
                <p className="max-w-3xl text-lg leading-8 text-neutral-300">
                  训练营参考并改编自《Go 语言圣经中文版》，用 GoGopher Arch 的实战语境重写知识点，并为每章配一个可运行的 sandbox 练习。
                </p>
              </div>
              <div className="flex flex-wrap gap-3">
                <Button asChild className="rounded-xl bg-[#00ADD8] px-6 py-5 font-bold text-neutral-950 hover:bg-[#00ADD8]/90">
                  <Link to={`/courses/go-basics/${goBasicsChapters[0].slug}`}>
                    开始第一章
                    <ArrowRight className="h-4 w-4" />
                  </Link>
                </Button>
                <Button asChild className="rounded-xl border border-neutral-700 bg-neutral-900 px-6 py-5 text-neutral-100 hover:bg-neutral-800">
                  <a href="https://github.com/gopl-zh/gopl-zh.github.com" target="_blank" rel="noreferrer">
                    查看原项目
                  </a>
                </Button>
              </div>
            </div>

            <Card className="border-neutral-800 bg-neutral-900/80 text-neutral-100 shadow-2xl shadow-[#00ADD8]/10">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <ShieldCheck className="h-5 w-5 text-[#00ADD8]" />
                  来源与改编说明
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4 text-sm leading-7 text-neutral-300">
                <p>本训练营不镜像原教程全文，课程导读、练习和验收标准由 GoGopher Arch 重写。</p>
                <p>
                  原教程正文与代码遵循其各自授权说明：仓库 LICENSE 为 BSD 3-Clause，附录 C 说明正文采用 CC-BY 3.0，代码遵循 Go 项目的 BSD 协议。
                </p>
                <div className="flex flex-wrap gap-2">
                  <Button asChild variant="outline" className="border-neutral-700 bg-neutral-950 text-neutral-200 hover:bg-neutral-800">
                    <a href="https://github.com/gopl-zh/gopl-zh.github.com/blob/master/LICENSE" target="_blank" rel="noreferrer">LICENSE</a>
                  </Button>
                  <Button asChild variant="outline" className="border-neutral-700 bg-neutral-950 text-neutral-200 hover:bg-neutral-800">
                    <a href="https://gopl-zh.github.io/appendix/appendix-c-cpoyright.html" target="_blank" rel="noreferrer">译文授权</a>
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      <section className="px-6 py-16">
        <div className="container mx-auto max-w-6xl space-y-8">
          <div className="flex flex-wrap items-end justify-between gap-4">
            <div>
              <h2 className="text-3xl font-bold text-white">课程路径</h2>
              <p className="mt-2 text-neutral-400">13 章按原书主题组织，每章先理解概念，再用 sandbox 跑一个小练习。</p>
            </div>
            <div className="flex gap-2 text-sm text-neutral-400">
              <span>章节数: {goBasicsChapters.length}</span>
              <span>·</span>
              <span>练习数: {goBasicsChapters.length}</span>
            </div>
          </div>

          {validationErrors.length > 0 && (
            <div className="rounded-xl border border-yellow-500/30 bg-yellow-500/10 p-4 text-sm text-yellow-200">
              课程数据校验提示：{validationErrors.join("；")}
            </div>
          )}

          <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
            {goBasicsChapters.map((chapter) => (
              <Link key={chapter.slug} to={`/courses/go-basics/${chapter.slug}`} className="group block">
                <Card className="h-full border-neutral-800 bg-neutral-900/70 text-neutral-100 transition-colors hover:border-[#00ADD8]/50 hover:bg-neutral-900">
                  <CardHeader className="space-y-4">
                    <div className="flex items-center justify-between gap-3">
                      <Badge className="border-neutral-700 bg-neutral-800 text-neutral-300">第 {chapter.order} 章</Badge>
                      <Badge className="border-[#00ADD8]/30 bg-[#00ADD8]/10 text-[#00ADD8]">{chapter.difficulty}</Badge>
                    </div>
                    <CardTitle className="flex items-center justify-between gap-4 text-xl text-white">
                      {chapter.title}
                      <ArrowRight className="h-5 w-5 text-neutral-500 transition-transform group-hover:translate-x-1 group-hover:text-[#00ADD8]" />
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-5">
                    <p className="line-clamp-3 text-sm leading-6 text-neutral-400">{chapter.summary}</p>
                    <div className="grid grid-cols-3 gap-2 text-xs text-neutral-400">
                      <CourseMetric icon={<Clock3 className="h-4 w-4" />} label={chapter.duration} />
                      <CourseMetric icon={<GraduationCap className="h-4 w-4" />} label={`${chapter.goals.length} 个目标`} />
                      <CourseMetric icon={<BookOpen className="h-4 w-4" />} label="1 个练习" />
                    </div>
                    <div className="flex items-center gap-2 text-sm text-green-300">
                      <CheckCircle2 className="h-4 w-4" />
                      重写导读 + sandbox 练习
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}

function CourseMetric({ icon, label }: { icon: React.ReactNode; label: string }) {
  return (
    <div className="flex flex-col items-center gap-1 rounded-lg border border-neutral-800 bg-neutral-950 px-2 py-3 text-center">
      <span className="text-[#00ADD8]">{icon}</span>
      <span>{label}</span>
    </div>
  );
}
