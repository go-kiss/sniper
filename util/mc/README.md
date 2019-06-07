# mc

memcache 基础库，开放有限接口，支持日志、opentracing 和 prometheus 监控。

# 配置

MC 配置，格式为 MC_${NAME}_HOSTS = "host1"，通过 ${NAME} 可以获取 MC 连接池。初始连接数使用 MC_DEFAULT_INIT_CONNS，最大连接数 MC_DEFAULT_MAX_IDLE_CONNS。

MC_XXX_HOSTS 只能填一个 memcache 实例。

高可用架构请使用 [twemcache](https://github.com/twitter/twemcache) 等中间件。

# 示例
```go
import "context"
import "sniper/util/mc"

ctx := context.Background()
c := mc.Get(ctx, "default")

err := c.Delete(ctx, "foo")
```
