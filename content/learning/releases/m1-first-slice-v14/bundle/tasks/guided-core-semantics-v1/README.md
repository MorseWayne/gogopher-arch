# 用 Go 语义整理名称

补全 `ClassifyNames`：去掉两端空白，拒绝空名称或超过 `maxRunes` 个 Unicode code point 的名称；接受项转换为 `Label`，并累计接受名称的 rune 数。`maxRunes <= 0` 时全部拒绝。

不要用字节数代替 rune 数。保留 `Summary` 的零值作为空结果，不需要额外构造器。
