# 字节帧所有权复验

`DetachPayload` 跳过 `headerBytes` 后返回独立 payload；负数或超过长度时返回 nil。`Frequency` 返回每个 byte 的出现次数，空输入也返回非 nil 空 map。

这是集合语义的异题复验，不提供实现提示。
