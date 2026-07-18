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
  'bounded-pipeline-practiced': '受控并发映射通过',
  'bounded-concurrency-respected': '并发数量保持在约束内',
  'pipeline-completes-without-leak': '流水线完整结束且未遗留任务',
  'shared-state-practiced': '共享计数状态得到同步',
  'registry-contract-correct': '状态快照契约正确',
  'race-detector-clean': '真实竞态检测通过',
  'synchronization-choice-explained': '能够解释同步方案取舍',
  'cancellation-practiced': '取消信号能够终止并发任务',
  'cancellation-propagates': '取消和失败能够传递到同批任务',
  'workers-release-before-return': '函数返回前已释放全部 Worker',
  'consumer-interface-practiced': '使用方接口调用符合契约',
  'consumer-interface-minimal': '使用方接口保持最小边界',
  'generic-helper-declared': '复用函数已使用类型参数',
  'substitute-contract-correct': '测试替身可以代入并保留错误',
  'generic-helper-reusable': '泛型算法可跨类型复用',
  'failure-triaged': '已从失败现场定位并修复缺陷',
  'regression-fixed': '回归行为已经修复',
  'static-analysis-clean': '静态检查通过',
  'profile-diagnosis-explained': '已说明性能证据与修复关系',
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
