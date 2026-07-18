import { Home, LayoutDashboard, RefreshCw, TriangleAlert } from 'lucide-react'
import { Link, isRouteErrorResponse, useRouteError } from 'react-router'

import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'

const dynamicImportErrorPatterns = [
  /failed to fetch dynamically imported module/i,
  /error loading dynamically imported module/i,
  /importing a module script failed/i,
  /loading chunk .* failed/i,
  /chunkloaderror/i,
]

export function isDynamicImportError(error: unknown) {
  const message = error instanceof Error
    ? `${error.name}: ${error.message}`
    : String(error)

  return dynamicImportErrorPatterns.some((pattern) => pattern.test(message))
}

export function RouteError() {
  const error = useRouteError()
  const needsFreshBuild = isDynamicImportError(error)

  const title = needsFreshBuild ? '应用已更新' : '页面暂时无法打开'
  const description = needsFreshBuild
    ? '当前页面引用了旧版本资源。重新加载后会同步最新版本，已经保存的学习进度不会丢失。'
    : routeErrorDescription(error)

  return (
    <main className="grid min-h-svh place-items-center bg-muted/25 px-4">
      <Card className="w-full max-w-xl">
        <CardHeader>
          <TriangleAlert className="mb-3 size-8 text-primary" />
          <CardTitle>{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 sm:flex-row">
          <Button onClick={() => window.location.reload()}>
            <RefreshCw />
            重新加载最新版本
          </Button>
          <Button asChild variant="outline">
            <Link to="/dashboard">
              <LayoutDashboard />
              学习工作台
            </Link>
          </Button>
          <Button asChild variant="ghost">
            <Link to="/">
              <Home />
              首页
            </Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  )
}

function routeErrorDescription(error: unknown) {
  if (isRouteErrorResponse(error)) {
    return `请求失败（${error.status}）。请重新加载；如果问题仍然存在，请稍后再试。`
  }

  return '页面加载时遇到了问题。请重新加载；如果问题仍然存在，请稍后再试。'
}
