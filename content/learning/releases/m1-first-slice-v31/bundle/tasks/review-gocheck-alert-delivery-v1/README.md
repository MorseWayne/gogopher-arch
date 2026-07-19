# 变式复验：alert delivery client

实现 `Client.Deliver`：向 `/v1/deliveries` POST `{"destination":"...","message":"..."}`，携带调用方 Context、Content-Type 与 Accept。202 时限制读取并解析 `delivery_id`。

429、其他 4xx、5xx 和超限响应分别映射为已有领域错误；每次调用只发送一次请求，并始终关闭响应 Body。补充至少五个命名 case。
