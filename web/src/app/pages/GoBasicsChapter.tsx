import type { ReactNode } from "react";
import { Link, useParams } from "react-router";
import { ArrowLeft, ArrowRight, BookOpen, CheckCircle2, Clock3, Lightbulb, MapPin, Star, Target } from "lucide-react";
import { CourseExercisePanel } from "../components/CourseExercisePanel";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { getGoBasicsChapterBySlug, getRelatedMissions, goBasicsChapters, type GoCourseChapter } from "../data/goBasicsCourse";

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

  return (
    <main className="flex-1 bg-neutral-950 text-neutral-50">
      <div className="container mx-auto max-w-6xl px-6 py-8 md:py-12">
        <div className="mb-8 flex items-center justify-between gap-4 border-b border-neutral-800 pb-4">
          <Link to="/courses/go-basics" className="inline-flex items-center gap-2 text-sm text-neutral-400 transition-colors hover:text-white">
            <ArrowLeft className="h-4 w-4" />
            返回训练营
          </Link>
          <div className="flex items-center gap-2">
            <ChapterNavButton chapter={previousChapter} label="上一章" direction="previous" />
            <ChapterNavButton chapter={nextChapter} label="下一章" direction="next" />
          </div>
        </div>

        <section className="mb-8 space-y-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Badge className="border-neutral-700 bg-neutral-800 text-neutral-300">第 {chapter.order} 章 · Go 基础训练营</Badge>
            <Badge className="border-[#00ADD8]/30 bg-[#00ADD8]/10 text-[#00ADD8]">{chapter.difficulty}</Badge>
          </div>

          <div className="space-y-4">
            <h1 className="text-3xl font-bold tracking-tight text-white md:text-5xl">{chapter.title}</h1>
            <p className="max-w-3xl text-lg leading-8 text-neutral-300">{chapter.summary}</p>
            <div className="flex flex-wrap gap-4 text-sm text-neutral-400">
              <ChapterMeta icon={<Clock3 className="h-4 w-4" />} text={chapter.duration} />
              <ChapterMeta icon={<Star className="h-4 w-4" />} text={chapter.difficulty} />
              <ChapterMeta icon={<BookOpen className="h-4 w-4" />} text={`${chapter.lessons.length} 个课程小节`} />
            </div>
          </div>
        </section>

        <section className="grid gap-6 lg:grid-cols-[minmax(0,3fr)_minmax(320px,2fr)]">
          <div className="space-y-6">
            <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <Target className="h-5 w-5 text-[#00ADD8]" />
                  学习目标
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="grid gap-3 md:grid-cols-2">
                  {chapter.goals.map((goal) => (
                    <li key={goal} className="flex gap-3 rounded-xl border border-neutral-800 bg-neutral-950 p-4 text-sm leading-6 text-neutral-300">
                      <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-green-400" />
                      {goal}
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>

            <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <BookOpen className="h-5 w-5 text-[#00ADD8]" />
                  课程正文
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-5">
                {chapter.lessons.map((lesson, index) => (
                  <article key={lesson.title} className="rounded-xl border border-neutral-800 bg-neutral-950 p-5">
                    <div className="mb-3 flex items-center gap-3">
                      <Badge className="border-neutral-700 bg-neutral-900 text-neutral-300">{index + 1}</Badge>
                      <h2 className="text-lg font-semibold text-white">{lesson.title}</h2>
                    </div>
                    <div className="space-y-3 text-sm leading-7 text-neutral-300">
                      {lesson.body.map((paragraph) => (
                        <p key={paragraph}>{paragraph}</p>
                      ))}
                    </div>
                    {lesson.code && (
                      <pre className="mt-4 overflow-auto rounded-xl border border-neutral-800 bg-black p-4 text-sm leading-6 text-neutral-200">
                        <code>{lesson.code}</code>
                      </pre>
                    )}
                  </article>
                ))}
              </CardContent>
            </Card>

            <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <Star className="h-5 w-5 text-[#00ADD8]" />
                  Go 1.24+ 与现代生态说明
                </CardTitle>
              </CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-2">
                {chapter.modernNotes.map((note) => (
                  <div key={note.title} className="rounded-xl border border-[#00ADD8]/20 bg-[#00ADD8]/5 p-4">
                    <h3 className="mb-2 font-semibold text-white">{note.title}</h3>
                    <p className="text-sm leading-6 text-neutral-300">{note.body}</p>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <CheckCircle2 className="h-5 w-5 text-[#00ADD8]" />
                  后端工程实践
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-3">
                  {chapter.engineeringPractices.map((practice) => (
                    <li key={practice} className="flex gap-3 rounded-xl border border-neutral-800 bg-neutral-950 p-4 text-sm leading-6 text-neutral-300">
                      <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-green-400" />
                      {practice}
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>

            <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <Lightbulb className="h-5 w-5 text-yellow-300" />
                  常见坑
                </CardTitle>
              </CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-3">
                {chapter.pitfalls.map((pitfall) => (
                  <div key={pitfall.title} className="rounded-xl border border-yellow-500/20 bg-yellow-500/5 p-4">
                    <h3 className="mb-2 font-semibold text-white">{pitfall.title}</h3>
                    <p className="mb-3 text-sm leading-6 text-yellow-100/80">现象：{pitfall.symptom}</p>
                    <p className="text-sm leading-6 text-neutral-300">修正：{pitfall.fix}</p>
                  </div>
                ))}
              </CardContent>
            </Card>

            <CourseExercisePanel chapterSlug={chapter.slug} exercise={chapter.exercise} />
          </div>

          <aside className="space-y-6">
            <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
              <CardHeader>
                <CardTitle className="text-white">验收 checklist</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-3">
                  {chapter.checklist.map((item) => (
                    <li key={item} className="flex gap-3 text-sm leading-6 text-neutral-300">
                      <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-green-400" />
                      {item}
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>

            <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <BookOpen className="h-5 w-5 text-[#00ADD8]" />
                  复盘问题
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ol className="space-y-3">
                  {chapter.reviewQuestions.map((question, index) => (
                    <li key={question} className="rounded-xl border border-neutral-800 bg-neutral-950 p-4 text-sm leading-6 text-neutral-300">
                      <span className="mr-2 font-mono text-[#00ADD8]">Q{index + 1}</span>
                      {question}
                    </li>
                  ))}
                </ol>
              </CardContent>
            </Card>

            <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-white">
                  <MapPin className="h-5 w-5 text-[#00ADD8]" />
                  衔接实习任务
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {relatedMissions.length > 0 ? (
                  relatedMissions.map((mission) => (
                    <Link key={mission.slug} to={`/missions/${mission.slug}`} className="block rounded-xl border border-neutral-800 bg-neutral-950 p-4 transition-colors hover:border-[#00ADD8]/50">
                      <div className="mb-1 text-sm font-semibold text-white">{mission.title}</div>
                      <div className="text-xs text-neutral-500">{mission.chapter} · {mission.difficulty}</div>
                    </Link>
                  ))
                ) : (
                  <div className="rounded-xl border border-neutral-800 bg-neutral-950 p-4 text-sm text-neutral-400">暂无绑定实习任务。</div>
                )}
              </CardContent>
            </Card>
          </aside>
        </section>
      </div>
    </main>
  );
}

function ChapterNotFound() {
  return (
    <main className="flex-1 bg-neutral-950 px-6 py-24 text-neutral-50">
      <div className="container mx-auto max-w-2xl rounded-2xl border border-neutral-800 bg-neutral-900/70 p-8 text-center">
        <Badge className="mb-4 border-yellow-500/30 bg-yellow-500/10 text-yellow-300">未找到章节</Badge>
        <h1 className="mb-3 text-3xl font-bold text-white">这个 Go 基础章节不存在</h1>
        <p className="mb-6 text-neutral-400">请返回训练营总览，从 13 章课程路径中选择一个有效章节。</p>
        <Button asChild className="rounded-xl bg-[#00ADD8] font-bold text-neutral-950 hover:bg-[#00ADD8]/90">
          <Link to="/courses/go-basics">返回训练营总览</Link>
        </Button>
      </div>
    </main>
  );
}

function ChapterNavButton({ chapter, label, direction }: { chapter?: GoCourseChapter; label: string; direction: "previous" | "next" }) {
  const content = (
    <>
      {direction === "previous" && <ArrowLeft className="h-4 w-4" />}
      {label}
      {direction === "next" && <ArrowRight className="h-4 w-4" />}
    </>
  );

  if (!chapter) {
    return <Button disabled className="h-9 rounded-lg border border-neutral-800 bg-neutral-900 px-3 text-neutral-700">{content}</Button>;
  }

  return (
    <Button asChild className="h-9 rounded-lg border border-neutral-800 bg-neutral-900 px-3 text-neutral-300 hover:bg-neutral-800 hover:text-white">
      <Link to={`/courses/go-basics/${chapter.slug}`}>{content}</Link>
    </Button>
  );
}

function ChapterMeta({ icon, text }: { icon: ReactNode; text: string }) {
  return (
    <span className="inline-flex items-center gap-2">
      {icon}
      {text}
    </span>
  );
}
