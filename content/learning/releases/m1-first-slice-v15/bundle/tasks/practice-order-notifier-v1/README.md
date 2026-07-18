# 用使用方接口建立测试接缝

补全 `Service.Place`。`Service` 只需要发送一条通知，因此接口只表达 `Notify`，不要求实现方暴露配置、关闭或重试等额外能力。

要求：

- 通知消息为 `order placed: <item>`；
- 把 notifier 返回的错误原样返回；
- 测试替身必须能观察 user、message 和调用次数；
- 不要在业务方法里创建具体邮件客户端。
