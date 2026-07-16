type CapabilityLabelSource = {
  id: string
  name: string
  description: string
}

const capabilityLabels: Record<string, Pick<CapabilityLabelSource, 'name' | 'description'>> = {
  'M1-01': {
    name: '使用 Go 工具链获取反馈',
    description: '运行 Build、Test、Vet，并判断下一步应该检查代码、测试还是运行环境。',
  },
}

const ruleLabels: Record<string, string> = {
  'module-builds': '代码可以完成编译',
  'toolchain-checks-pass': '工具链检查通过',
  'tool-feedback-explained': '理解并解释工具反馈',
  'toolchain-baseline-builds': '代码可以完成编译',
  'toolchain-reflection-recorded': '已写下工具反馈小结',
  'error-chain-preserved': '错误原因得到保留',
  'resource-closed': '资源得到正确关闭',
  'invalid-input-rejected': '无效输入得到拒绝',
  'stable-output': '输出结果保持稳定',
  'cli-failure-contract': '命令失败行为符合预期',
  'learner-tests-present': '补充了自己的测试',
  'visible-tests-pass': '公开测试通过',
  'held-out-tests-pass': '最终检查通过',
}

export function learnerCapabilityName(capability?: CapabilityLabelSource): string {
  if (!capability) return '学习目标'
  return capabilityLabels[capability.id]?.name ?? capability.name
}

export function learnerCapabilityDescription(capability: CapabilityLabelSource): string {
  return capabilityLabels[capability.id]?.description ?? capability.description
}

export function learnerRuleLabel(ruleID: string): string {
  return ruleLabels[ruleID] ?? '练习要求检查'
}
