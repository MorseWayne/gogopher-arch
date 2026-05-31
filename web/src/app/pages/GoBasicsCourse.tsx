import type { ReactNode } from "react";
import { Link } from "react-router";
import { ArrowRight, BookOpen, CheckCircle2, Clock3, GraduationCap, ShieldCheck, Star } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Progress } from "../components/ui/progress";
import { goBasicsChapters, getGoBasicsExerciseCount, getGoBasicsLessonCount, validateGoBasicsCourse } from "../data/goBasicsCourse";

const validationErrors = validateGoBasicsCourse();
const lessonCount = goBasicsChapters.reduce((sum, chapter) => sum + getGoBasicsLessonCount(chapter), 0);
const modernNoteCount = goBasicsChapters.reduce((sum, chapter) => sum + chapter.modernNotes.length, 0);
const exerciseCount = goBasicsChapters.reduce((sum, chapter) => sum + getGoBasicsExerciseCount(chapter), 0);
const mdxChapterCount = goBasicsChapters.filter((chapter) => chapter.contentKind === "mdx").length;
const firstChapter = goBasicsChapters[0];

export function GoBasicsCourse() {
  return (
    <main className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-6 md:px-6 md:py-8">
      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
        <Card className="border-primary/20 bg-background shadow-sm">
          <CardHeader className="gap-4">
            <Badge className="w-fit">Go 基础训练营 · 13 章内置课程</Badge>
            <div>
              <CardTitle className="text-3xl md:text-5xl">从第一行 Go 代码到后端实习基本功</CardTitle>
              <CardDescription className="mt-4 max-w-3xl text-base leading-7">
                GoGopher Arch 内置课程开始迁移到 React + MDX 内容系统。课程文章用 Markdown/MDX 管理，页面继续保留工程实践、常见坑、复盘问题和可运行 sandbox 练习。
              </CardDescription>
            </div>
            <div className="flex flex-wrap gap-3">
              {firstChapter && (
                <Button asChild>
                  <Link to={`/courses/go-basics/${firstChapter.slug}`}>
                    开始第一章
                    <ArrowRight data-icon="inline-end" />
                  </Link>
                </Button>
              )}
              <Button asChild variant="outline">
                <Link to="/dashboard">返回学习总览</Link>
              </Button>
            </div>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ShieldCheck className="text-primary" />
              内置课程说明
            </CardTitle>
            <CardDescription>课程正文迁移到 MDX 后，仍在应用内完成练习和验收。</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2">
            <HeroMetric label="章节" value={`${goBasicsChapters.length}`} />
            <HeroMetric label="正文小节" value={`${lessonCount}`} />
            <HeroMetric label="现代说明" value={`${modernNoteCount}`} />
            <HeroMetric label="练习" value={`${exerciseCount}`} />
            <HeroMetric label="MDX 章节" value={`${mdxChapterCount}`} />
          </CardContent>
        </Card>
      </section>

      {validationErrors.length > 0 && (
        <Alert>
          <ShieldCheck />
          <AlertTitle>课程数据完整性提示</AlertTitle>
          <AlertDescription>{validationErrors.join("；")}</AlertDescription>
        </Alert>
      )}

      <section className="grid gap-4 lg:grid-cols-[260px_minmax(0,1fr)]">
        <aside className="h-fit lg:sticky lg:top-20">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">学习建议</CardTitle>
              <CardDescription>先读正文，再运行练习，最后进入任务线。</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <div>
                <div className="mb-2 flex justify-between text-sm">
                  <span className="text-muted-foreground">访客演示进度</span>
                  <span>18%</span>
                </div>
                <Progress value={18} />
              </div>
              <ul className="flex flex-col gap-3 text-sm text-muted-foreground">
                <li className="flex gap-2"><CheckCircle2 className="text-primary" /> 每章都配一个可运行练习。</li>
                <li className="flex gap-2"><CheckCircle2 className="text-primary" /> 课程正文优先服务浏览器练习。</li>
                <li className="flex gap-2"><CheckCircle2 className="text-primary" /> 完成基础后进入后端实习任务线。</li>
              </ul>
            </CardContent>
          </Card>
        </aside>

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {goBasicsChapters.map((chapter) => (
            <Link key={chapter.slug} to={`/courses/go-basics/${chapter.slug}`} className="group">
              <Card className="h-full transition-colors hover:border-primary/40">
                <CardHeader>
                  <div className="flex items-center justify-between gap-3">
                    <Badge variant="outline">第 {chapter.order} 章</Badge>
                    <Badge variant="secondary">{chapter.difficulty}</Badge>
                  </div>
                  <CardTitle className="flex items-center justify-between gap-3 text-xl">
                    {chapter.title}
                    <ArrowRight className="size-5 text-muted-foreground transition-transform group-hover:translate-x-1 group-hover:text-primary" />
                  </CardTitle>
                  <CardDescription className="line-clamp-3 leading-6">{chapter.summary}</CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col gap-4">
                  <div className="grid grid-cols-2 gap-2 text-sm text-muted-foreground">
                    <CourseMetric icon={<Clock3 />} label={chapter.duration} />
                    <CourseMetric icon={<BookOpen />} label={`${getGoBasicsLessonCount(chapter)} 节`} />
                    <CourseMetric icon={<Star />} label={chapter.contentKind === "mdx" ? "MDX 文章" : `${chapter.modernNotes.length} 说明`} />
                    <CourseMetric icon={<GraduationCap />} label={`${getGoBasicsExerciseCount(chapter)} 练习`} />
                  </div>
                  <div className="flex items-center gap-2 text-sm text-primary">
                    <CheckCircle2 className="size-4" />
                    {chapter.contentKind === "mdx" ? "MDX 文章 + sandbox 练习" : "内置正文 + sandbox 练习"}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      </section>
    </main>
  );
}

function HeroMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border bg-muted/30 p-4">
      <div className="text-2xl font-bold">{value}</div>
      <div className="text-sm text-muted-foreground">{label}</div>
    </div>
  );
}

function CourseMetric({ icon, label }: { icon: ReactNode; label: string }) {
  return (
    <div className="flex items-center gap-2 rounded-xl border bg-background p-3">
      <span className="text-primary">{icon}</span>
      <span>{label}</span>
    </div>
  );
}
