# 变式复验：alert acknowledgement 事务

完成 `Store.Acknowledge`。在 serializable transaction 中先登记 `(rule_id,idempotency_key)`；重复 key 读取并返回既有 rule。首次命令用 expected version 条件更新 acknowledged_by 和 version，无匹配行返回 `ErrConflict`。

所有查询必须使用调用方 Context。错误或冲突回滚，成功或重复重放提交。补充至少四个命名测试 case。
