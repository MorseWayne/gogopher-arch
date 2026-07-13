import { Link } from 'react-router'
import { ArrowLeft, LayoutDashboard, RouteOff } from 'lucide-react'

import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'

export function NotFound() {
  return (
    <main className="grid min-h-svh place-items-center bg-muted/25 px-4">
      <Card className="w-full max-w-xl">
        <CardHeader>
          <RouteOff className="mb-3 size-8 text-primary" />
          <CardTitle>页面不存在</CardTitle>
          <CardDescription>
            当前产品只暴露服务端驱动的 Dashboard 与 Activity route，不提供静态 Course、Mission 或 Sandbox fallback。
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 sm:flex-row">
          <Button asChild>
            <Link to="/dashboard">
              <LayoutDashboard />
              前往学习工作台
            </Link>
          </Button>
          <Button asChild variant="outline">
            <Link to="/">
              <ArrowLeft />
              返回首页
            </Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  )
}
