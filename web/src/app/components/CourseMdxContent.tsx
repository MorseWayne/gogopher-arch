import type { ComponentProps, ComponentType } from "react";
import type { MDXComponents } from "mdx/types";
import { cn } from "./ui/utils";

type CourseMdxComponent = ComponentType<{ components?: MDXComponents }>;

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
};

export function CourseMdxContent({ Content }: { Content: CourseMdxComponent }) {
  return (
    <div className="space-y-5 rounded-2xl border bg-background p-5">
      <Content components={mdxComponents} />
    </div>
  );
}
