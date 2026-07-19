# 输出格式 module 复验

独立完成 `formatmodule`。`render.Lines` 把每个 `Record` 格式化为 `<key>=<value>` 并保持输入顺序；`Document.Count` 返回行数；`cmd/render` 只依赖导出 API。

`render/document.go` 的 package comment 以 `Package render` 开头；所有导出类型、函数和方法都写以声明名开头的 doc comment。
