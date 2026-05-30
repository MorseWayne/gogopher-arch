import type { ReactNode } from "react";
import { Link } from "react-router";
import { ArrowRight, BookOpen, CheckCircle2, Clock3, GraduationCap, ShieldCheck, Star } from "lucide-react";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { goBasicsChapters, validateGoBasicsCourse } from "../data/goBasicsCourse";

const validationErrors = validateGoBasicsCourse();
const lessonCount = goBasicsChapters.reduce((sum, chapter) => sum + chapter.lessons.length, 0);
const modernNoteCount = goBasicsChapters.reduce((sum, chapter) => sum + chapter.modernNotes.length, 0);

export function GoBasicsCourse() {
  return (
    <main className="flex-1 bg-neutral-950 text-neutral-50">
      <section className="border-b border-neutral-900 bg-[radial-gradient(circle_at_top,_rgba(0,173,216,0.16),_transparent_40%),#0a0a0a] px-6 py-20">
        <div className="container mx-auto max-w-6xl">
          <div className="grid gap-10 lg:grid-cols-[minmax(0,3fr)_minmax(320px,2fr)] lg:items-center">
            <div className="space-y-6">
              <Badge className="border-[#00ADD8]/30 bg-[#00ADD8]/10 text-[#00ADD8]">Go 基础训练营 · 13 章内置课程</Badge>
              <div className="space-y-4">
                <h1 className="text-4xl font-extrabold tracking-tight text-white md:text-6xl">从第一行 Go 代码到后端实习基本功</h1>
                <p className="max-w-3xl text-lg leading-8 text-neutral-300">
                  GoGopher Arch 内置课程，结合 Go 1.24+ 和后端工程现状重制。13 章覆盖语法、数据结构、测试、并发和底层边界，每章都配有可运行 sandbox 练习。
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
                  <Link to="/dashboard">进入实习工作台</Link>
                </Button>
              </div>
            </div>

            <Card className="border-neutral-800 bg-neutral-900/80 text-neutral-100 shadow-2xl shadow-[#00ADD8]/10">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <ShieldCheck className="h-5 w-5 text-[#00ADD8]" />
                  内置课程说明
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-5 text-sm leading-7 text-neutral-300">
                <p>课程讲解、练习、验收 checklist 和复盘问题由 GoGopher Arch 重新整理生成，目标是服务浏览器练习和后端实习任务衔接。</p>
                <div className="grid grid-cols-2 gap-3">
                  <HeroMetric label="章节" value={`${goBasicsChapters.length}`} />
                  <HeroMetric label="正文小节" value={`${lessonCount}`} />
                  <HeroMetric label="现代说明" value={`${modernNoteCount}`} />
                  <HeroMetric label="练习" value={`${goBasicsChapters.length}`} />
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
              <p className="mt-2 text-neutral-400">13 章按工程成长顺序组织，每章先学习内置正文，再用 sandbox 完成一个小练习。</p>
            </div>
            <div className="flex gap-2 text-sm text-neutral-400">
              <span>章节数: {goBasicsChapters.length}</span>
              <span>·</span>
              <span>练习数: {goBasicsChapters.length}</span>
            </div>
          </div>

          {validationErrors.length > 0 && (
            <div className="rounded-xl border border-yellow-500/30 bg-yellow-500/10 p-4 text-sm text-yellow-200">
              内置课程数据完整性提示：{validationErrors.join("；")}
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
                    <div className="grid grid-cols-4 gap-2 text-xs text-neutral-400">
                      <CourseMetric icon={<Clock3 className="h-4 w-4" />} label={chapter.duration} />
                      <CourseMetric icon={<BookOpen className="h-4 w-4" />} label={`${chapter.lessons.length} 节`} />
                      <CourseMetric icon={<Star className="h-4 w-4" />} label={`${chapter.modernNotes.length} 说明`} />
                      <CourseMetric icon={<GraduationCap className="h-4 w-4" />} label="1 练习" />
                    </div>
                    <div className="flex items-center gap-2 text-sm text-green-300">
                      <CheckCircle2 className="h-4 w-4" />
                      内置正文 + 工程实践 + sandbox 练习
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

function HeroMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-neutral-800 bg-neutral-950 p-4">
      <div className="text-2xl font-bold text-white">{value}</div>
      <div className="text-xs text-neutral-500">{label}</div>
    </div>
  );
}

function CourseMetric({ icon, label }: { icon: ReactNode; label: string }) {
  return (
    <div className="flex flex-col items-center gap-1 rounded-lg border border-neutral-800 bg-neutral-950 px-2 py-3 text-center">
      <span className="text-[#00ADD8]">{icon}</span>
      <span>{label}</span>
    </div>
  );
}
