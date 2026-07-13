# 配置合并器变式复验

实现 `mergecfg`：读取 base 和 override JSON，以 service 名称合并，校验 endpoint 与 retry，并按名称稳定输出。该任务不复用 assessment 的字段结构。
