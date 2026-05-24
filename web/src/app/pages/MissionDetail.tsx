import type { ReactNode } from "react";
import { Link, useParams } from "react-router";
import {
  ArrowLeft,
  ArrowRight,
  BookOpen,
  CheckCircle2,
  CheckSquare,
  Clock3,
  Code2,
  GraduationCap,
  Lightbulb,
  Lock,
  MapPin,
  Square,
  Star,
  Target,
} from "lucide-react";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "../components/ui/accordion";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Separator } from "../components/ui/separator";
import { goFunctions, goKeywords, missionCatalog, missions, statusMeta, type Mission } from "../data/missions";

export function MissionDetail() {
  const { slug } = useParams();
  const requestedMission = slug ? missionCatalog[slug] : undefined;
  const mission = requestedMission ?? missions[0];
  const missionIndex = missions.findIndex((item) => item.slug === mission.slug);
  const previousMission = missions[missionIndex - 1];
  const nextMission = missions[missionIndex + 1];
  const completedCriteria = mission.status === "completed" ? mission.criteria.length : mission.status === "in-progress" ? Math.min(2, mission.criteria.length) : 0;

  return (
    <main className="flex-1 bg-neutral-950 text-neutral-50">
      <div className="container mx-auto px-6 py-8 md:py-12">
        <div className="mb-8 flex items-center justify-between gap-4 border-b border-neutral-800 pb-4">
          <Link to="/#roadmap" className="inline-flex items-center gap-2 text-sm text-neutral-400 transition-colors hover:text-white">
            <ArrowLeft className="h-4 w-4" />
            返回路线图
          </Link>
          <div className="flex items-center gap-2">
            <MissionNavButton mission={previousMission} label="上一关" direction="previous" />
            <MissionNavButton mission={nextMission} label="下一关" direction="next" />
          </div>
        </div>

        <section className="mb-8 space-y-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Badge className="border-neutral-700 bg-neutral-800 text-neutral-300">
              Day {mission.day} · {mission.chapter}
            </Badge>
            <Badge className={statusMeta[mission.status].className}>{statusMeta[mission.status].label}</Badge>
          </div>

          {!requestedMission && slug && (
            <div className="rounded-xl border border-yellow-500/30 bg-yellow-500/5 px-4 py-3 text-sm text-yellow-300">
              未找到指定任务，已展示当前可继续的任务。
            </div>
          )}

          <div className="space-y-4">
            <h1 className="text-3xl font-bold tracking-tight text-white md:text-4xl">{mission.title}</h1>
            <div className="flex flex-wrap gap-4 text-sm text-neutral-400">
              <MissionMeta icon={<Clock3 className="h-4 w-4" />} text={mission.duration} />
              <MissionMeta icon={<Star className="h-4 w-4" />} text={mission.difficulty} />
              <MissionMeta icon={<MapPin className="h-4 w-4" />} text={`前置: ${mission.prerequisite.join("、")}`} />
            </div>
          </div>
        </section>

        <section className="grid gap-6 lg:grid-cols-[minmax(0,3fr)_minmax(320px,2fr)]">
          <Card className="border-neutral-800 bg-neutral-900/70 text-neutral-100">
            <CardContent className="px-6 py-2">
              <Accordion type="multiple" defaultValue={["background", "objectives"]}>
                <AccordionItem value="background" className="border-neutral-800">
                  <AccordionTrigger className="text-base text-white hover:text-[#00ADD8] hover:no-underline">
                    <SectionTitle icon={<BookOpen className="h-5 w-5" />} title="任务背景" />
                  </AccordionTrigger>
                  <AccordionContent className="space-y-3 text-neutral-300">
                    {mission.background.map((paragraph) => (
                      <p key={paragraph} className="leading-7">
                        {paragraph}
                      </p>
                    ))}
                  </AccordionContent>
                </AccordionItem>

                <AccordionItem value="objectives" className="border-neutral-800">
                  <AccordionTrigger className="text-base text-white hover:text-[#00ADD8] hover:no-underline">
                    <SectionTitle icon={<Target className="h-5 w-5" />} title="任务目标" />
                  </AccordionTrigger>
                  <AccordionContent>
                    <ul className="space-y-3 text-neutral-300">
                      {mission.objectives.map((objective) => (
                        <li key={objective} className="flex gap-3 leading-6">
                          <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-[#00ADD8]" />
                          <span>{objective}</span>
                        </li>
                      ))}
                    </ul>
                  </AccordionContent>
                </AccordionItem>

                <AccordionItem value="hints" className="border-neutral-800">
                  <AccordionTrigger className="text-base text-white hover:text-[#00ADD8] hover:no-underline">
                    <SectionTitle icon={<Lightbulb className="h-5 w-5" />} title="关键提示" />
                  </AccordionTrigger>
                  <AccordionContent>
                    <div className="space-y-3 rounded-xl border border-yellow-500/30 bg-yellow-500/5 p-4 text-sm text-yellow-100">
                      {mission.hints.map((hint) => (
                        <p key={hint} className="leading-6">
                          {hint}
                        </p>
                      ))}
                    </div>
                  </AccordionContent>
                </AccordionItem>

                <AccordionItem value="knowledge" className="border-neutral-800">
                  <AccordionTrigger className="text-base text-white hover:text-[#00ADD8] hover:no-underline">
                    <SectionTitle icon={<GraduationCap className="h-5 w-5" />} title="前置知识" />
                  </AccordionTrigger>
                  <AccordionContent>
                    <div className="flex flex-wrap gap-2">
                      {mission.knowledge.map((item) => (
                        <Badge key={item} className="border-neutral-700 bg-neutral-950 text-neutral-300">
                          {item}
                        </Badge>
                      ))}
                    </div>
                  </AccordionContent>
                </AccordionItem>
              </Accordion>
            </CardContent>
          </Card>

          <Card className="h-fit border-neutral-800 bg-neutral-900 text-neutral-100 lg:sticky lg:top-24">
            <CardHeader>
              <CardTitle className="text-lg font-bold text-white">验收标准</CardTitle>
            </CardHeader>
            <CardContent className="space-y-5">
              <ul className="space-y-3">
                {mission.criteria.map((criterion, index) => {
                  const done = index < completedCriteria;
                  return (
                    <li key={criterion} className="flex items-start gap-3 text-sm text-neutral-300">
                      {done ? <CheckSquare className="mt-0.5 h-4 w-4 shrink-0 text-green-400" /> : <Square className="mt-0.5 h-4 w-4 shrink-0 text-neutral-600" />}
                      <span>{criterion}</span>
                    </li>
                  );
                })}
              </ul>
              <Separator className="bg-neutral-800" />
              <MissionCta mission={mission} />
            </CardContent>
          </Card>
        </section>

        <section className="mt-8 space-y-3">
          <div className="flex items-center gap-2 text-sm font-semibold text-neutral-400">
            <Code2 className="h-4 w-4" />
            初始代码
          </div>
          <CodePreview code={mission.starterCode} locked={mission.status === "locked"} />
        </section>
      </div>
    </main>
  );
}

function MissionMeta({ icon, text }: { icon: ReactNode; text: string }) {
  return (
    <span className="inline-flex items-center gap-2">
      {icon}
      {text}
    </span>
  );
}

function SectionTitle({ icon, title }: { icon: ReactNode; title: string }) {
  return (
    <span className="inline-flex items-center gap-2">
      <span className="text-[#00ADD8]">{icon}</span>
      {title}
    </span>
  );
}

function MissionNavButton({ mission, label, direction }: { mission?: Mission; label: string; direction: "previous" | "next" }) {
  const content = (
    <>
      {direction === "previous" && <ArrowLeft className="h-4 w-4" />}
      {label}
      {direction === "next" && <ArrowRight className="h-4 w-4" />}
    </>
  );

  if (!mission) {
    return (
      <Button disabled className="h-9 rounded-lg border border-neutral-800 bg-neutral-900 px-3 text-neutral-700">
        {content}
      </Button>
    );
  }

  return (
    <Button asChild className="h-9 rounded-lg border border-neutral-800 bg-neutral-900 px-3 text-neutral-300 hover:bg-neutral-800 hover:text-white">
      <Link to={`/missions/${mission.slug}`}>{content}</Link>
    </Button>
  );
}

function MissionCta({ mission }: { mission: Mission }) {
  if (mission.status === "locked") {
    return (
      <Button disabled className="w-full rounded-xl bg-neutral-800 py-4 font-bold text-neutral-500">
        <Lock className="h-4 w-4" />
        完成前置任务解锁
      </Button>
    );
  }

  const label = mission.status === "completed" ? "查看复盘" : mission.status === "in-progress" ? "继续挑战" : "开始挑战";
  const className = mission.status === "completed"
    ? "w-full rounded-xl border border-white/20 bg-transparent py-4 font-bold text-white hover:bg-white/5"
    : "w-full rounded-xl bg-[#00ADD8] py-4 font-bold text-neutral-950 hover:bg-[#00ADD8]/90";

  return (
    <Button asChild className={className}>
      <Link to={`/dashboard?mission=${mission.slug}`}>{label}</Link>
    </Button>
  );
}

function CodePreview({ code, locked }: { code: string; locked: boolean }) {
  const lines = code.split("\n");

  return (
    <div className="relative overflow-hidden rounded-xl border border-neutral-800 bg-[#0d0d0d]">
      {locked && <div className="absolute inset-0 z-10 flex items-center justify-center bg-neutral-950/60 text-sm font-semibold text-neutral-400 backdrop-blur-[2px]">完成前置任务后解锁代码</div>}
      <div className={locked ? "overflow-x-auto blur-sm" : "overflow-x-auto"}>
        <div className="grid min-w-[720px] grid-cols-[auto_1fr] text-sm font-mono leading-relaxed">
          <div className="select-none border-r border-neutral-800 bg-neutral-950 px-4 py-6 text-right text-neutral-600">
            {lines.map((_, index) => (
              <div key={index + 1}>{index + 1}</div>
            ))}
          </div>
          <pre className="p-6 text-neutral-300">
            {lines.map((line, index) => (
              <div key={`${index}-${line}`}>{renderGoLine(line)}</div>
            ))}
          </pre>
        </div>
      </div>
    </div>
  );
}

function renderGoLine(line: string) {
  const commentStart = line.indexOf("//");
  const code = commentStart >= 0 ? line.slice(0, commentStart) : line;
  const comment = commentStart >= 0 ? line.slice(commentStart) : "";

  return (
    <>
      {renderGoTokens(code)}
      {comment && <span className="text-neutral-500 italic">{comment}</span>}
    </>
  );
}

function renderGoTokens(code: string) {
  return code.split(/("(?:\\.|[^"\\])*"|`[^`]*`|\b\d+\b|\b[A-Za-z_][A-Za-z0-9_]*\b|\s+|.)/g).filter(Boolean).map((token, index) => {
    if (/^\s+$/.test(token)) {
      return token;
    }

    if (/^"/.test(token) || /^`/.test(token)) {
      return <span key={`${token}-${index}`} className="text-green-400">{token}</span>;
    }

    if (/^\d+$/.test(token)) {
      return <span key={`${token}-${index}`} className="text-purple-400">{token}</span>;
    }

    if (goKeywords.has(token)) {
      return <span key={`${token}-${index}`} className="text-pink-500">{token}</span>;
    }

    if (goFunctions.has(token)) {
      return <span key={`${token}-${index}`} className="text-blue-400">{token}</span>;
    }

    return token;
  });
}
