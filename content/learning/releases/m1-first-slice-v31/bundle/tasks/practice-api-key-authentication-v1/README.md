# API key 认证练习

完成 `authn.New` 与 `Authenticate`：校验配置，只保留 token digest，严格解析 Bearer header，并以恒定时间比较返回对应 subject。
