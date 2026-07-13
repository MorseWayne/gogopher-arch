import { Link, Outlet } from 'react-router'
import { Github, Home, LayoutDashboard, Terminal } from 'lucide-react'

import { Button } from './ui/button'
import { ThemeToggle } from './ThemeToggle'

export function PublicLayout() {
  return (
    <div className="flex min-h-svh flex-col bg-background text-foreground selection:bg-primary selection:text-primary-foreground">
      <header className="sticky top-0 z-50 border-b bg-background/85 backdrop-blur-xl">
        <div className="mx-auto flex h-16 w-full max-w-7xl items-center justify-between px-4 md:px-6">
          <Brand />
          <nav className="flex items-center gap-2">
            <Button asChild variant="ghost" size="sm" className="hidden sm:inline-flex">
              <a href="https://github.com/MorseWayne/gogopher-arch" target="_blank" rel="noreferrer">
                <Github />
                GitHub
              </a>
            </Button>
            <ThemeToggle />
            <Button asChild size="sm">
              <Link to="/dashboard">
                <LayoutDashboard />
                学习工作台
              </Link>
            </Button>
          </nav>
        </div>
      </header>

      <main className="flex-1">
        <Outlet />
      </main>

      <footer className="border-t bg-muted/30">
        <div className="mx-auto flex w-full max-w-7xl flex-col gap-3 px-4 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between md:px-6">
          <span>© 2026 GoGopher Arch · MIT License</span>
          <div className="flex items-center gap-4">
            <Link to="/dashboard" className="hover:text-foreground">学习工作台</Link>
            <a
              href="https://github.com/MorseWayne/gogopher-arch"
              target="_blank"
              rel="noreferrer"
              className="hover:text-foreground"
            >
              Source
            </a>
          </div>
        </div>
      </footer>
    </div>
  )
}

export function LearningLayout() {
  return (
    <div className="min-h-svh bg-muted/25 text-foreground">
      <header className="sticky top-0 z-40 border-b bg-background/90 backdrop-blur-xl">
        <div className="mx-auto flex h-16 w-full max-w-7xl items-center gap-3 px-4 md:px-6">
          <Link to="/" aria-label="GoGopher Arch 首页" className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground">
            <Terminal className="size-5" />
          </Link>
          <div className="min-w-0">
            <div className="truncate font-semibold">Capability Evidence Lab</div>
            <div className="truncate text-xs text-muted-foreground">服务端事实驱动的学习工作台</div>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <Button asChild variant="ghost" size="sm" className="hidden sm:inline-flex">
              <Link to="/">
                <Home />
                首页
              </Link>
            </Button>
            <ThemeToggle />
          </div>
        </div>
      </header>
      <Outlet />
    </div>
  )
}

function Brand() {
  return (
    <Link to="/" className="flex items-center gap-2 font-bold">
      <span className="flex size-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
        <Terminal className="size-5" />
      </span>
      <span>GoGopher Arch</span>
    </Link>
  )
}
