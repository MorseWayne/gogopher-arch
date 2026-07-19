# 练习：受控 HTTP client

实现 `NewClient`。拒绝非正数配置；Clone 默认 Transport，设置 MaxIdleConns、MaxIdleConnsPerHost、IdleConnTimeout、TLSHandshakeTimeout 和 ResponseHeaderTimeout，再返回带整体 Timeout 的独立 `http.Client`。
