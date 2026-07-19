# 独立评估：run 完成事务

完成 `Store.Complete`。验证输入后以 `sql.LevelSerializable` 开始事务，先向 `run_commands(run_id,idempotency_key)` 插入命令标记并使用 `ON CONFLICT ... DO NOTHING RETURNING run_id`。无返回行表示重复 key，此时在同一事务读取既有 run 并提交。

首次命令用 `WHERE id=$2 AND version=$3` 条件更新 status 和 version，并 RETURNING 完整结果；无返回行映射为 `ErrConflict`。所有错误与冲突必须回滚，只有完整成功或重复读取成功才能提交。补充至少四个命名测试 case。
