# 亲手完成第一个 Go 程序

这节课先写代码，再使用工具。请补全 `welcome`，让它返回 `welcome, <name>`，然后让 `main` 把结果输出到终端。

## 你将完成

1. 阅读 `package main`、`import`、`func` 和 `return` 所在的位置。
2. 补全 `welcome`，不要把测试中的名字写死。
3. 运行 **Test**，根据失败信息修改代码。
4. 再运行 **Build** 和 **Vet**，最后提交。

## 完成标准

- `welcome("Gopher")` 返回 `welcome, Gopher`；
- 换成其他名字时仍能得到正确结果；
- Test、Build 和 Vet 均能通过。

不要修改 `go.mod` 和测试文件。这里的目标不是背命令，而是第一次走完“写代码—运行—读反馈—修正”的循环。
