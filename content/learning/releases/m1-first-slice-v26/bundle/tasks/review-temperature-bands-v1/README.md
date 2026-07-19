# 温度区间复验

实现 `ClassifyReadings`。当 `low >= high` 时返回零值；否则逐项生成 `Reading`，小于 `low` 为 `BandLow`，大于 `high` 为 `BandHigh`，其余为 `BandNormal`，并累计非正常项到 `Alerts`。

这是异题复验，不提供提示。请独立使用命名类型、零值、循环和分支完成。
