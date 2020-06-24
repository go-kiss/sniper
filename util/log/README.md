# log

log 目前最低级别是 debug，可以通过 LOG_LEVEL 环境变量或者配置项指定。

log 会记录上下文信息，所以需要传入一个 ctx 才能获取 log 实例。

## 示例
```go
import "sniper/util/log"

log.Get(ctx).Errorf("1 + 2 = %d", 1 + 2)
```
