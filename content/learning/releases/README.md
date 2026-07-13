# Learning releases

每个子目录是由 `cmd/learning-content release` 生成的不可变内容发布，运行时只能从这些发布包加载定义和任务资产。

此目录的 `go.mod` 仅用于把归档的 Go 测试资产隔离在应用 module 之外；它不属于任何 release，也不会进入 release bundle hash。
