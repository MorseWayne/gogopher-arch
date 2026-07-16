import type { ComponentType, ReactNode } from 'react'
import type { MDXComponents } from 'mdx/types'
import { BookOpenText, CircleAlert } from 'lucide-react'

import { CourseMdxContent } from '../CourseMdxContent'

type LearningModule = {
  default: ComponentType<{ components?: MDXComponents }>
}

const contentModules = import.meta.glob<LearningModule>('../../../content/learning/*.mdx', { eager: true })

export function LearningContent({ contentRef }: { contentRef?: string }) {
  const modulePath = normalizeContentRef(contentRef)
  const Content = modulePath ? contentModules[modulePath]?.default : undefined

  if (!contentRef) return null

  return (
    <section id="lesson-content" aria-labelledby="lesson-content-title" className="scroll-mt-6 space-y-4">
      <div>
        <div className="flex items-center gap-2 text-sm font-semibold text-primary">
          <BookOpenText className="size-4" />课程讲解
        </div>
        <h2 id="lesson-content-title" className="mt-1 text-2xl font-bold tracking-tight">先理解，再动手</h2>
      </div>
      {Content ? (
        <CourseMdxContent Content={Content} />
      ) : (
        <ContentState icon={<CircleAlert />} text="这份课程内容暂时不可用，请稍后重试。" />
      )}
    </section>
  )
}

function normalizeContentRef(contentRef?: string): string {
  if (!contentRef) return ''
  if (contentRef.startsWith('web/src/content/')) {
    return `../../../content/${contentRef.slice('web/src/content/'.length)}`
  }
  if (contentRef.startsWith('/src/content/')) {
    return `../../../content/${contentRef.slice('/src/content/'.length)}`
  }
  return ''
}

function ContentState({ icon, text }: { icon: ReactNode; text: string }) {
  return <div className="flex items-center gap-3 rounded-2xl border bg-muted/30 p-5 text-sm text-muted-foreground">{icon}{text}</div>
}
