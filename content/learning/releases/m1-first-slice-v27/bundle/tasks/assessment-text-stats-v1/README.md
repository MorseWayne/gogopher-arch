# 文本记录分类器

独立实现 `Analyze`。逐条保留原始字符串并按 Unicode code point 统计长度：空字符串为 `CategoryEmpty`；`maxRunes <= 0` 或长度超过上限为 `CategoryTooLong`；其余为 `CategoryAccepted`。只累计接受项的数量和 rune 总数，其余计入 `Rejected`。

空输入应自然返回 `Report` 零值。不要修改已给出的类型和常量契约。
