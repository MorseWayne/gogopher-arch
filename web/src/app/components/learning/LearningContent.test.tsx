import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { LearningContent } from './LearningContent'

describe('LearningContent', () => {
  it.each([
    ['guided', 'web/src/content/learning/core-semantics-guided-v1.mdx', '先理解，再动手', '从“语法认识”走到“语义能用”'],
    ['practice', 'web/src/content/learning/slice-ownership-practice-v1.mdx', '先确认契约，再动手', 'slice 是视图，所有权要由契约说清'],
    ['practice', 'web/src/content/learning/error-contract-practice-v1.mdx', '先确认契约，再动手', '错误要同时服务人和程序'],
    ['practice', 'web/src/content/learning/json-normalization-practice-v1.mdx', '先确认契约，再动手', '规范化不是“顺手排个序”'],
    ['practice', 'web/src/content/learning/table-tests-practice-v1.mdx', '先确认契约，再动手', '让失败输出直接说清行为'],
    ['assessment', 'web/src/content/learning/config-normalizer-assessment-v1.mdx', '独立完成，按契约验收', '把四项能力组合成一个命令'],
    ['assessment', 'web/src/content/learning/status-module-assessment-v1.mdx', '独立完成，按契约验收', '独立交付一个可被使用的 module'],
    ['review', 'web/src/content/learning/config-merge-review-v1.mdx', '换一个情境，重新证明掌握', '在新情境中重新证明掌握'],
    ['review', 'web/src/content/learning/format-module-review-v1.mdx', '换一个情境，重新证明掌握', '从状态报告迁移到输出格式包'],
  ])('renders %s-specific framing and content', (mode, contentRef, frameTitle, contentTitle) => {
    const { unmount } = render(<LearningContent mode={mode} contentRef={contentRef} />)

    expect(screen.getByRole('heading', { name: frameTitle })).toBeVisible()
    expect(screen.getByRole('heading', { name: contentTitle })).toBeVisible()
    unmount()
  })

  it('keeps a visible unavailable state for an unknown content reference', () => {
    render(<LearningContent mode="practice" contentRef="web/src/content/learning/missing.mdx" />)

    expect(screen.getByText('这份课程内容暂时不可用，请稍后重试。')).toBeVisible()
  })
})
