# 变式复习：指标文本生成故障

使用 `testdata` 中的失败摘要、breakpoint 记录和 `alloc_space` profile 修复 `Render`：不能遗漏第一项，输出格式为 `<name>:<value>\n`，并用 `strings.Builder` 避免重复拼接。

同时修复 Vet 能发现的格式化参数错误。提交小结不少于 80 个字符，并包含 `breakpoint`、`alloc_space`、`strings.Builder`，说明证据到修改之间的推理。
