import type { ReactNode } from "react";
import { Fragment } from "react";
import { Outlet, Link, useLocation } from "react-router";
import {
  BookOpen,
  Bot,
  Boxes,
  ChevronRight,
  Code2,
  ExternalLink,
  Github,
  Home,
  LayoutDashboard,
  Rocket,
  Terminal,
} from "lucide-react";
import { Badge } from "./ui/badge";
import { ThemeToggle } from "./ThemeToggle";
import { Button } from "./ui/button";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "./ui/breadcrumb";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
} from "./ui/sidebar";
import { goBasicsChapters } from "../data/goBasicsCourse";
import { missions } from "../data/missions";

const firstChapter = goBasicsChapters[0];
const firstMission = missions[0];

type SandboxTarget = {
  href: string;
  label: string;
  reason?: string;
};

export function PublicLayout() {
  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground selection:bg-primary selection:text-primary-foreground">
      <header className="sticky top-0 z-50 border-b bg-background/85 backdrop-blur-xl">
        <div className="mx-auto flex h-16 w-full max-w-7xl items-center justify-between px-4 md:px-6">
          <Link to="/" className="flex items-center gap-2 font-bold text-foreground">
            <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
              <Terminal />
            </span>
            <span>GoGopher Arch</span>
          </Link>

          <nav className="hidden items-center gap-1 md:flex">
            <Button asChild variant="ghost" size="sm">
              <Link to="/">首页</Link>
            </Button>
            <Button asChild variant="ghost" size="sm">
              <Link to="/courses/go-basics">Go 基础训练营</Link>
            </Button>
            <Button asChild variant="ghost" size="sm">
              <Link to="/#paths">学习路径</Link>
            </Button>
            <Button asChild variant="ghost" size="sm">
              <Link to="/dashboard">学习总览</Link>
            </Button>
          </nav>

          <div className="flex items-center gap-2">
            <ThemeToggle />
            <Button asChild variant="outline" size="sm" className="hidden md:inline-flex">
              <a href="https://github.com/MorseWayne/gogopher-arch" target="_blank" rel="noreferrer">
                <Github data-icon="inline-start" />
                GitHub
              </a>
            </Button>
            <Button asChild size="sm">
              <Link to="/dashboard">
                进入学习区
                <ChevronRight data-icon="inline-end" />
              </Link>
            </Button>
          </div>
        </div>
      </header>

      <main className="flex-1">
        <Outlet />
      </main>

      <footer className="border-t bg-muted/30">
        <div className="mx-auto flex w-full max-w-7xl flex-col gap-4 px-4 py-8 text-sm text-muted-foreground md:flex-row md:items-center md:justify-between md:px-6">
          <div className="flex items-center gap-2">
            <Terminal className="size-4" />
            <span>© 2026 GoGopher Arch. MIT License.</span>
          </div>
          <div className="flex flex-wrap gap-4">
            <Link to="/courses/go-basics" className="hover:text-foreground">Go 基础训练营</Link>
            <Link to="/dashboard" className="hover:text-foreground">学习总览</Link>
            <a href="https://github.com/MorseWayne/gogopher-arch" target="_blank" rel="noreferrer" className="inline-flex items-center gap-1 hover:text-foreground">
              GitHub
              <ExternalLink className="size-3" />
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}

export function LearningLayout() {
  const location = useLocation();
  const sandboxTarget = resolveSandboxTarget(location.pathname);
  const breadcrumbs = getBreadcrumbs(location.pathname);

  return (
    <SidebarProvider>
      <LearningSidebar sandboxTarget={sandboxTarget} pathname={location.pathname} />
      <SidebarInset>
        <header className="sticky top-0 z-40 flex h-14 items-center gap-3 border-b bg-background/90 px-4 backdrop-blur-xl md:px-6">
          <SidebarTrigger />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink asChild>
                  <Link to="/dashboard">学习区</Link>
                </BreadcrumbLink>
              </BreadcrumbItem>
              {breadcrumbs.map((item) => (
                <Fragment key={item.label}>
                  <BreadcrumbSeparator />
                  <BreadcrumbItem>
                    {item.href ? (
                      <BreadcrumbLink asChild>
                        <Link to={item.href}>{item.label}</Link>
                      </BreadcrumbLink>
                    ) : (
                      <BreadcrumbPage>{item.label}</BreadcrumbPage>
                    )}
                  </BreadcrumbItem>
                </Fragment>
              ))}
            </BreadcrumbList>
          </Breadcrumb>
          <div className="ml-auto flex items-center gap-2">
            <ThemeToggle />
            <div className="hidden items-center gap-2 text-sm text-muted-foreground md:flex">
              <Badge variant="secondary">访客本地会话</Badge>
              <Button asChild variant="outline" size="sm">
                <Link to={sandboxTarget.href}>
                  <Terminal data-icon="inline-start" />
                  沙盒
                </Link>
              </Button>
            </div>
          </div>
        </header>
        <div className="min-h-[calc(100svh-3.5rem)] bg-muted/30">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

function LearningSidebar({ sandboxTarget, pathname }: { sandboxTarget: SandboxTarget; pathname: string }) {
  return (
    <Sidebar collapsible="offcanvas">
      <SidebarHeader className="gap-3 border-b p-4">
        <Link to="/" className="flex items-center gap-3">
          <span className="flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
            <Terminal />
          </span>
          <span className="min-w-0">
            <span className="block truncate font-bold">GoGopher Arch</span>
            <span className="block truncate text-xs text-muted-foreground">Go 后端实习成长平台</span>
          </span>
        </Link>
        <div className="rounded-xl border bg-background p-3">
          <div className="mb-2 flex items-center justify-between text-xs text-muted-foreground">
            <span>访客本地会话</span>
            <Badge variant="outline">Lv.1</Badge>
          </div>
          <div className="h-2 rounded-full bg-muted">
            <div className="h-2 w-[18%] rounded-full bg-primary" />
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>工作区</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarNavLink href="/dashboard" active={pathname === "/dashboard"} icon={<LayoutDashboard />} label="总览" />
              <SidebarNavLink href={sandboxTarget.href} active={false} icon={<Terminal />} label="沙盒快捷入口" badge="跳转" />
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarSeparator />

        <SidebarGroup>
          <SidebarGroupLabel>学习路径</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarNavLink href="/courses/go-basics" active={pathname.startsWith("/courses/go-basics")} icon={<BookOpen />} label="Go 基础训练营" />
              <SidebarNavLink href={`/missions/${firstMission.slug}`} active={pathname.startsWith("/missions")} icon={<Code2 />} label="后端实习任务线" />
              <SidebarComingSoon icon={<Boxes />} label="工程能力进阶" />
              <SidebarComingSoon icon={<Bot />} label="AI 全栈路线" />
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarSeparator />

        <SidebarGroup>
          <SidebarGroupLabel>项目</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarNavLink href="/" active={false} icon={<Home />} label="公开首页" />
              <SidebarExternalLink href="https://github.com/MorseWayne/gogopher-arch" icon={<Github />} label="GitHub / README" />
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t p-4">
        <div className="rounded-xl bg-muted p-3 text-xs text-muted-foreground">
          <div className="mb-1 flex items-center gap-2 font-medium text-foreground">
            <Rocket className="size-4 text-primary" />
            下一步建议
          </div>
          先完成 Go 基础训练营，再进入后端实习任务线。
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}

function SidebarNavLink({ href, active, icon, label, badge }: { href: string; active: boolean; icon: ReactNode; label: string; badge?: string }) {
  return (
    <SidebarMenuItem>
      <SidebarMenuButton asChild isActive={active} tooltip={label}>
        <Link to={href}>
          {icon}
          <span>{label}</span>
        </Link>
      </SidebarMenuButton>
      {badge && <SidebarMenuBadge>{badge}</SidebarMenuBadge>}
    </SidebarMenuItem>
  );
}

function SidebarExternalLink({ href, icon, label }: { href: string; icon: ReactNode; label: string }) {
  return (
    <SidebarMenuItem>
      <SidebarMenuButton asChild tooltip={label}>
        <a href={href} target="_blank" rel="noreferrer">
          {icon}
          <span>{label}</span>
        </a>
      </SidebarMenuButton>
    </SidebarMenuItem>
  );
}

function SidebarComingSoon({ icon, label }: { icon: ReactNode; label: string }) {
  return (
    <SidebarMenuItem>
      <SidebarMenuButton disabled tooltip={`${label}即将开放`}>
        {icon}
        <span>{label}</span>
      </SidebarMenuButton>
      <SidebarMenuBadge>Soon</SidebarMenuBadge>
    </SidebarMenuItem>
  );
}

function resolveSandboxTarget(pathname: string): SandboxTarget {
  if (pathname.startsWith("/missions/")) {
    return { href: `${pathname}#sandbox`, label: "当前任务沙盒" };
  }

  if (pathname.startsWith("/courses/go-basics/") && pathname !== "/courses/go-basics") {
    return { href: `${pathname}#exercise`, label: "当前章节练习" };
  }

  if (firstChapter) {
    return {
      href: `/courses/go-basics/${firstChapter.slug}#exercise`,
      label: "默认 Go 练习",
      reason: "当前页面没有沙盒，跳转到第一个可运行章节。",
    };
  }

  return { href: "/courses/go-basics", label: "课程总览", reason: "暂无可运行练习，返回课程总览。" };
}

function getBreadcrumbs(pathname: string) {
  if (pathname === "/dashboard") {
    return [{ label: "总览" }];
  }

  if (pathname === "/courses/go-basics") {
    return [{ label: "Go 基础训练营" }];
  }

  if (pathname.startsWith("/courses/go-basics/")) {
    const slug = pathname.split("/").at(-1);
    const chapter = goBasicsChapters.find((item) => item.slug === slug);
    return [{ label: "Go 基础训练营", href: "/courses/go-basics" }, { label: chapter ? `第 ${chapter.order} 章` : "章节" }];
  }

  if (pathname.startsWith("/missions/")) {
    const slug = pathname.split("/").at(-1);
    const mission = missions.find((item) => item.slug === slug);
    return [{ label: "后端实习任务线" }, { label: mission?.title ?? "任务" }];
  }

  return [{ label: "学习区" }];
}
