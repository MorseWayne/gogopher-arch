import type { ComponentType, ReactNode } from "react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router";
import { ArrowLeft, ArrowRight, BookOpen, CheckCircle2, Clock3, Lightbulb, MapPin, Star, Target } from "lucide-react";
import { CourseExercisePanel } from "../components/CourseExercisePanel";
import { CourseMdxContent } from "../components/CourseMdxContent";
import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Separator } from "../components/ui/separator";
import { getGoBasicsChapterBySlug, getGoBasicsLessonCount, getRelatedMissions, goBasicsChapters, type GoCourseChapter } from "../data/goBasicsCourse";

export function GoBasicsChapter() {
  const { chapterSlug } = useParams();
  const chapter = getGoBasicsChapterBySlug(chapterSlug);

  if (!chapter) {
    return <ChapterNotFound />;
  }

  const chapterIndex = goBasicsChapters.findIndex((item) => item.slug === chapter.slug);
  const previousChapter = goBasicsChapters[chapterIndex - 1];
  const nextChapter = goBasicsChapters[chapterIndex + 1];
  const relatedMissions = getRelatedMissions(chapter);
  const hasMdxContent = Boolean(chapter.loadContent);
  const headings = ["学习目标", hasMdxContent ? "课程文章" : "课程正文", "现代生态说明", "工程实践", "常见坑", "沙盒练习", "复盘问题"];

  return (
    <main className="mx-auto grid w-full max-w-[112rem] gap-6 px-4 py-6 md:px-6 md:py-8 2xl:px-8 xl:grid-cols-[220px_minmax(0,1fr)_240px] 2xl:grid-cols-[240px_minmax(0,1fr)_260px]">
      <aside className="hidden xl:block">
        <Card className="sticky top-20">
          <CardHeader>
            <CardTitle className="text-lg">章节目录</CardTitle>
            <CardDescription>Go 基础训练营</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-1">
            {goBasicsChapters.map((item) => (
              <Button key={item.slug} asChild variant={item.slug === chapter.slug ? "secondary" : "ghost"} className="justify-start">
                <Link to={`/courses/go-basics/${item.slug}`}>第 {item.order} 章 · {item.title}</Link>
              </Button>
            ))}
          </CardContent>
        </Card>
      </aside>

      <article className="min-w-0 space-y-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Button asChild variant="ghost" size="sm">
            <Link to="/courses/go-basics">
              <ArrowLeft data-icon="inline-start" />
              返回训练营
            </Link>
          </Button>
          <div className="flex items-center gap-2">
            <ChapterNavButton chapter={previousChapter} label="上一章" direction="previous" />
            <ChapterNavButton chapter={nextChapter} label="下一章" direction="next" />
          </div>
        </div>

        <Card className="border-primary/20">
          <CardHeader className="gap-4">
            <div className="flex flex-wrap gap-2">
              <Badge>第 {chapter.order} 章</Badge>
              <Badge variant="secondary">{chapter.difficulty}</Badge>
              <Badge variant="outline">{chapter.duration}</Badge>
            </div>
            <div>
              <CardTitle className="text-3xl md:text-5xl">{chapter.title}</CardTitle>
              <CardDescription className="mt-4 text-base leading-7">{chapter.summary}</CardDescription>
            </div>
            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <ChapterMeta icon={<Clock3 />} text={chapter.duration} />
              <ChapterMeta icon={<Star />} text={chapter.difficulty} />
              <ChapterMeta icon={<BookOpen />} text={`${getGoBasicsLessonCount(chapter)} 个课程小节`} />
            </div>
          </CardHeader>
        </Card>

        <Card id="goals">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Target className="text-primary" />
              学习目标
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="grid gap-3 md:grid-cols-2">
              {chapter.goals.map((goal) => (
                <li key={goal} className="flex gap-3 rounded-2xl border bg-muted/30 p-4 text-sm leading-6">
                  <CheckCircle2 className="mt-0.5 shrink-0 text-primary" />
                  {goal}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card id="lesson">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BookOpen className="text-primary" />
              课程文章
            </CardTitle>
            <CardDescription>{hasMdxContent ? "正文来自 MDX 内容文件，进入章节后按需加载，保留产品内 sandbox 和任务衔接。" : "浅色阅读 surface，代码块保持深色以突出运行语境。"}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            {chapter.loadContent ? (
              <LazyCourseMdxContent chapter={chapter} />
            ) : chapter.lessons.length === 0 ? (
              <Alert>
                <BookOpen />
                <AlertTitle>暂无正文内容</AlertTitle>
                <AlertDescription>请返回课程总览选择其他章节。</AlertDescription>
              </Alert>
            ) : (
              chapter.lessons.map((lesson, index) => (
                <section key={lesson.title} className="rounded-2xl border bg-background p-5">
                  <div className="mb-3 flex items-center gap-3">
                    <Badge variant="outline">{index + 1}</Badge>
                    <h2 className="text-xl font-semibold">{lesson.title}</h2>
                  </div>
                  <div className="space-y-3 text-sm leading-7 text-muted-foreground">
                    {lesson.body.map((paragraph) => (
                      <p key={paragraph}>{paragraph}</p>
                    ))}
                  </div>
                  {lesson.code && <CodeBlock code={lesson.code} />}
                </section>
              ))
            )}
          </CardContent>
        </Card>

        <Card id="modern">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Star className="text-primary" />
              Go 1.24+ 与现代生态说明
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2">
            {chapter.modernNotes.map((note) => (
              <div key={note.title} className="rounded-2xl border bg-primary/5 p-4">
                <h3 className="mb-2 font-semibold">{note.title}</h3>
                <p className="text-sm leading-6 text-muted-foreground">{note.body}</p>
              </div>
            ))}
          </CardContent>
        </Card>

        <Card id="practice">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CheckCircle2 className="text-primary" />
              后端工程实践
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3">
              {chapter.engineeringPractices.map((practice) => (
                <li key={practice} className="flex gap-3 rounded-2xl border bg-muted/30 p-4 text-sm leading-6">
                  <CheckCircle2 className="mt-0.5 shrink-0 text-primary" />
                  {practice}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card id="pitfalls">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Lightbulb className="text-primary" />
              常见坑
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-3">
            {chapter.pitfalls.map((pitfall) => (
              <div key={pitfall.title} className="rounded-2xl border bg-amber-50 p-4 text-amber-950 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-100">
                <h3 className="mb-2 font-semibold">{pitfall.title}</h3>
                <p className="mb-3 text-sm leading-6">现象：{pitfall.symptom}</p>
                <p className="text-sm leading-6">修正：{pitfall.fix}</p>
              </div>
            ))}
          </CardContent>
        </Card>

        <Card id="review">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BookOpen className="text-primary" />
              复盘问题
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ol className="space-y-3">
              {chapter.reviewQuestions.map((question, index) => (
                <li key={question} className="rounded-2xl border bg-muted/30 p-4 text-sm leading-6 text-muted-foreground">
                  <span className="mr-2 font-mono text-primary">Q{index + 1}</span>
                  {question}
                </li>
              ))}
            </ol>
          </CardContent>
        </Card>

        <div id="exercise">
          <CourseExercisePanel chapterSlug={chapter.slug} exercise={chapter.exercise} />
        </div>
      </article>

      <aside className="space-y-4 xl:sticky xl:top-20 xl:h-fit">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">本页目录</CardTitle>
            <CardDescription>内容为空时不影响正文阅读。</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-1">
            {headings.map((heading) => (
              <Button key={heading} asChild variant="ghost" className="justify-start">
                <a href={`#${tocId(heading)}`}>{heading}</a>
              </Button>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">验收 checklist</CardTitle>
          </CardHeader>
          <CardContent>
            {chapter.checklist.length === 0 ? (
              <Alert>
                <AlertTitle>暂无验收标准</AlertTitle>
                <AlertDescription>继续阅读正文或进入练习。</AlertDescription>
              </Alert>
            ) : (
              <ul className="space-y-3">
                {chapter.checklist.map((item) => (
                  <li key={item} className="flex gap-3 text-sm leading-6 text-muted-foreground">
                    <CheckCircle2 className="mt-0.5 shrink-0 text-primary" />
                    {item}
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <MapPin className="text-primary" />
              衔接实习任务
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {relatedMissions.length > 0 ? (
              relatedMissions.map((mission) => (
                <Link key={mission.slug} to={`/missions/${mission.slug}`} className="block rounded-2xl border bg-muted/30 p-4 transition-colors hover:border-primary/40">
                  <div className="mb-1 text-sm font-semibold">{mission.title}</div>
                  <div className="text-xs text-muted-foreground">{mission.chapter} · {mission.difficulty}</div>
                </Link>
              ))
            ) : (
              <div className="rounded-2xl border border-dashed bg-muted/30 p-4 text-sm text-muted-foreground">暂无绑定实习任务。</div>
            )}
          </CardContent>
        </Card>
      </aside>
    </main>
  );
}

function LazyCourseMdxContent({ chapter }: { chapter: GoCourseChapter }) {
  const [Content, setContent] = useState<ComponentType | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    let isActive = true;
    setContent(null);
    setError("");

    if (!chapter.loadContent) {
      return () => {
        isActive = false;
      };
    }

    chapter.loadContent()
      .then((LoadedContent) => {
        if (isActive) {
          setContent(() => LoadedContent);
        }
      })
      .catch((loadError) => {
        if (isActive) {
          setError(loadError instanceof Error ? loadError.message : "课程文章加载失败");
        }
      });

    return () => {
      isActive = false;
    };
  }, [chapter]);

  if (error) {
    return (
      <Alert variant="destructive">
        <BookOpen />
        <AlertTitle>课程文章加载失败</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (!Content) {
    return (
      <div className="rounded-2xl border bg-muted/30 p-5 text-sm leading-6 text-muted-foreground">
        正在加载 MDX 课程文章…
      </div>
    );
  }

  return <CourseMdxContent Content={Content} />;
}

function ChapterNotFound() {
  return (
    <main className="mx-auto flex min-h-[calc(100svh-3.5rem)] w-full max-w-3xl items-center px-4 py-16 md:px-6">
      <Card className="w-full text-center">
        <CardHeader>
          <Badge variant="secondary" className="mx-auto w-fit">未找到章节</Badge>
          <CardTitle className="text-3xl">这个 Go 基础章节不存在</CardTitle>
          <CardDescription>请返回训练营总览，从 13 章课程路径中选择一个有效章节。</CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild>
            <Link to="/courses/go-basics">返回训练营总览</Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}

function ChapterNavButton({ chapter, label, direction }: { chapter?: GoCourseChapter; label: string; direction: "previous" | "next" }) {
  const content = (
    <>
      {direction === "previous" && <ArrowLeft data-icon="inline-start" />}
      {label}
      {direction === "next" && <ArrowRight data-icon="inline-end" />}
    </>
  );

  if (!chapter) {
    return <Button disabled variant="outline" size="sm">{content}</Button>;
  }

  return (
    <Button asChild variant="outline" size="sm">
      <Link to={`/courses/go-basics/${chapter.slug}`}>{content}</Link>
    </Button>
  );
}

function ChapterMeta({ icon, text }: { icon: ReactNode; text: string }) {
  return (
    <span className="inline-flex items-center gap-2">
      <span className="text-primary">{icon}</span>
      {text}
    </span>
  );
}

function CodeBlock({ code }: { code: string }) {
  return (
    <pre className="mt-4 overflow-auto rounded-2xl border bg-slate-950 p-4 text-sm leading-6 text-slate-100">
      <code>{code}</code>
    </pre>
  );
}

function tocId(heading: string) {
  const map: Record<string, string> = {
    学习目标: "goals",
    课程文章: "lesson",
    课程正文: "lesson",
    现代生态说明: "modern",
    工程实践: "practice",
    常见坑: "pitfalls",
    沙盒练习: "exercise",
    复盘问题: "review",
  };

  return map[heading] ?? "lesson";
}
