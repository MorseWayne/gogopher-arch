import type { ReactNode } from "react";
import { Link } from "react-router";
import { ArrowRight, BookOpen, CheckCircle2, Clock3, GraduationCap, ShieldCheck, Star } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
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
                13 章 Go 基础内容可按目录自由浏览。每章包含课程文章、工程实践、常见坑、复盘问题和章节练习，不受个性化学习进度限制。
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
            <CardDescription>课程目录负责系统学习，学习工作台负责可执行练习、验收与下一步推荐。</CardDescription>
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
              <CardDescription>按顺序学习，或直接打开当前需要查阅的章节。</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <ul className="flex flex-col gap-3 text-sm text-muted-foreground">
                <li className="flex gap-2"><CheckCircle2 className="text-primary" /> 13 章内容都可以直接打开。</li>
                <li className="flex gap-2"><CheckCircle2 className="text-primary" /> 每章都包含起始代码和验收目标。</li>
                <li className="flex gap-2"><CheckCircle2 className="text-primary" /> 需要在线运行和反馈时进入学习工作台。</li>
              </ul>
              <Button asChild variant="outline">
                <Link to="/dashboard">进入学习工作台</Link>
              </Button>
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
                    {chapter.contentKind === "mdx" ? "MDX 文章 + 章节练习" : "内置正文 + 章节练习"}
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
