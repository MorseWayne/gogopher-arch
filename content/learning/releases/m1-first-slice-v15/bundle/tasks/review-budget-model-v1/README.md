# 预算模型异题复验

独立完成 `Budget`。`NewBudget` 拒绝空白名称和小于 1 的额度并去除名称两端空白；`Spend` 拒绝非正金额或超额支出且失败时不修改状态；`Refund` 拒绝非正金额或超过已支出金额；`Remaining` 返回余额，`Label` 输出 `<name>:<remaining>`。
