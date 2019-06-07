# redis

redis 基础库，开放有限接口，支持日志、opentracing 和 prometheus 监控。

# 配置

redis 配置，格式为 REDIS_${NAME}_HOST = "host1"，通过 ${NAME} 可以获取 redis 连接池。初始连接数使用 REDIS_DEFAULT_INIT_CONNS，最大连接数 REDIS_DEFAULT_MAX_CONNS。

REDIS_XXX_HOST 只能填一个 redis 实例。

高可用架构请使用 [envoy](https://www.envoyproxy.io/) 等中间件。

# 示例
```go
import "context"
import "sniper/util/redis"

ctx := context.Background()
c := redis.Get(ctx, "default")

err := c.Del(ctx, "foo")
```
