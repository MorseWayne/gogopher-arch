import type { ComponentProps, ComponentType, ReactNode } from "react";
import type { MDXComponents } from "mdx/types";
import { cn } from "./ui/utils";

type CourseMdxComponent = ComponentType<{ components?: MDXComponents }>;

type SourceNoteProps = {
  source: string;
  title?: string;
  chapter?: string;
  href?: string;
  note: string;
};

type CompareNoteProps = {
  title: string;
  points: string[];
};

type ExamplePairProps = {
  title: string;
  leftTitle: string;
  rightTitle: string;
  left: string;
  right: string;
};

type PitfallCardProps = {
  title: string;
  symptom: string;
  fix: string;
};

type PracticeBridgeProps = {
  exercise: string;
  text: string;
  href?: string;
};

type DeepDiveProps = {
  title: string;
  children: ReactNode;
};

const mdxComponents: MDXComponents = {
  h1: ({ className, ...props }: ComponentProps<"h1">) => (
    <h1 className={cn("text-3xl font-bold tracking-tight md:text-4xl", className)} {...props} />
  ),
  h2: ({ className, ...props }: ComponentProps<"h2">) => (
    <h2 className={cn("scroll-mt-24 border-b pb-2 pt-4 text-2xl font-semibold tracking-tight", className)} {...props} />
  ),
  h3: ({ className, ...props }: ComponentProps<"h3">) => (
    <h3 className={cn("scroll-mt-24 pt-3 text-xl font-semibold", className)} {...props} />
  ),
  p: ({ className, ...props }: ComponentProps<"p">) => (
    <p className={cn("text-sm leading-7 text-muted-foreground md:text-base", className)} {...props} />
  ),
  a: ({ className, ...props }: ComponentProps<"a">) => (
    <a className={cn("font-medium text-primary underline underline-offset-4", className)} {...props} />
  ),
  ul: ({ className, ...props }: ComponentProps<"ul">) => (
    <ul className={cn("ml-5 list-disc space-y-2 text-sm leading-7 text-muted-foreground md:text-base", className)} {...props} />
  ),
  ol: ({ className, ...props }: ComponentProps<"ol">) => (
    <ol className={cn("ml-5 list-decimal space-y-2 text-sm leading-7 text-muted-foreground md:text-base", className)} {...props} />
  ),
  li: ({ className, ...props }: ComponentProps<"li">) => <li className={cn("pl-1", className)} {...props} />,
  blockquote: ({ className, ...props }: ComponentProps<"blockquote">) => (
    <blockquote className={cn("rounded-2xl border-l-4 border-primary bg-primary/5 px-5 py-4 text-muted-foreground", className)} {...props} />
  ),
  pre: ({ className, ...props }: ComponentProps<"pre">) => (
    <pre className={cn("overflow-auto rounded-2xl border border-slate-800 bg-slate-950 p-4 text-sm leading-6 text-slate-100 shadow-sm", className)} {...props} />
  ),
  code: ({ className, ...props }: ComponentProps<"code">) => {
    const isBlockCode = Boolean(className?.includes("language-") || className?.includes("hljs"));

    return (
      <code
        className={cn(
          isBlockCode ? "block bg-transparent p-0 font-mono text-inherit" : "rounded-md bg-muted px-1.5 py-0.5 font-mono text-sm text-foreground",
          className,
        )}
        {...props}
      />
    );
  },
  table: ({ className, ...props }: ComponentProps<"table">) => (
    <div className="overflow-x-auto rounded-2xl border">
      <table className={cn("w-full border-collapse text-sm", className)} {...props} />
    </div>
  ),
  th: ({ className, ...props }: ComponentProps<"th">) => (
    <th className={cn("border-b bg-muted/50 px-3 py-2 text-left font-semibold", className)} {...props} />
  ),
  td: ({ className, ...props }: ComponentProps<"td">) => (
    <td className={cn("border-b px-3 py-2 text-muted-foreground", className)} {...props} />
  ),
  SourceNote,
  CompareNote,
  ExamplePair,
  DeepDive,
  PitfallCard,
  PracticeBridge,
};

export function CourseMdxContent({ Content }: { Content: CourseMdxComponent }) {
  return (
    <div className="space-y-5 rounded-2xl border bg-background p-5">
      <Content components={mdxComponents} />
    </div>
  );
}

function SourceNote({ source, title, chapter, href, note }: SourceNoteProps) {
  const heading = [source, title, chapter].filter(Boolean).join(" · ");

  return (
    <aside className="rounded-2xl border border-primary/20 bg-primary/5 p-4">
      <div className="mb-2 text-xs font-semibold uppercase tracking-wider text-primary">知识来源</div>
      <div className="text-sm font-semibold text-foreground">
        {href ? (
          <a className="text-primary underline underline-offset-4" href={href} target="_blank" rel="noreferrer">
            {heading}
          </a>
        ) : (
          heading
        )}
      </div>
      <p className="mt-2 text-sm leading-6 text-muted-foreground">{note}</p>
    </aside>
  );
}

function CompareNote({ title, points }: CompareNoteProps) {
  return (
    <aside className="rounded-2xl border bg-muted/40 p-4">
      <div className="mb-3 text-sm font-semibold text-foreground">{title}</div>
      <ul className="ml-5 list-disc space-y-2 text-sm leading-6 text-muted-foreground">
        {points.map((point) => (
          <li key={point}>{point}</li>
        ))}
      </ul>
    </aside>
  );
}

function ExamplePair({ title, leftTitle, rightTitle, left, right }: ExamplePairProps) {
  return (
    <section className="rounded-2xl border bg-muted/20 p-4">
      <div className="mb-3 text-sm font-semibold text-foreground">{title}</div>
      <div className="grid gap-3 lg:grid-cols-2">
        <CodeExample title={leftTitle} code={left} />
        <CodeExample title={rightTitle} code={right} />
      </div>
    </section>
  );
}

function CodeExample({ title, code }: { title: string; code: string }) {
  return (
    <div className="overflow-hidden rounded-2xl border border-slate-800 bg-slate-950">
      <div className="border-b border-slate-800 bg-slate-900 px-4 py-2 text-xs font-semibold uppercase tracking-wider text-slate-400">{title}</div>
      <pre className="overflow-auto p-4 text-sm leading-6 text-slate-100">
        <code>{code.trim()}</code>
      </pre>
    </div>
  );
}

function PitfallCard({ title, symptom, fix }: PitfallCardProps) {
  return (
    <aside className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-amber-950 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-100">
      <div className="mb-2 text-sm font-semibold">常见坑：{title}</div>
      <p className="text-sm leading-6">现象：{symptom}</p>
      <p className="mt-2 text-sm leading-6">修正：{fix}</p>
    </aside>
  );
}

function PracticeBridge({ exercise, text, href = "#exercise" }: PracticeBridgeProps) {
  return (
    <aside className="flex flex-col gap-3 rounded-2xl border border-cyan-900 bg-cyan-950 p-4 text-cyan-100 md:flex-row md:items-center md:justify-between">
      <div>
        <div className="mb-1 text-xs font-semibold uppercase tracking-wider text-cyan-300">练习衔接 · {exercise}</div>
        <p className="text-sm leading-6">{text}</p>
      </div>
      <a className="shrink-0 rounded-full bg-cyan-300 px-4 py-2 text-sm font-semibold text-cyan-950 transition hover:bg-cyan-200" href={href}>
        进入练习
      </a>
    </aside>
  );
}

function DeepDive({ title, children }: DeepDiveProps) {
  return (
    <details className="rounded-2xl border bg-muted/30 p-4">
      <summary className="cursor-pointer text-sm font-semibold text-foreground">深入理解：{title}</summary>
      <div className="mt-4 space-y-4">{children}</div>
    </details>
  );
}
