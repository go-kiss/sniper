# memdb

memdb 主要解决以下问题：

- 加载 redis 配置
- 记录 redis 执行日志
- 上报 opentracing 追踪数据
- 汇总 prometheus 监控指标

核心思想是用`github.com/ngrok/sqlmw`把现有的`database/sql`驱动包起来，
拦截所有数据库操作进行观察。

## 配置

框架默认只支持 redis。

每个数据库的配置需要指定一个名字，并添加`MEMDB_DSN_`前缀。

配置内容使用 url 格式，参数使用 query 字符串传递。

```yaml
MEMDB_DSN_BAR = "redis://name:password@localhost:6379?DB=1"
```

除了 hostname 之外，其他参数均为可选。支持的配置所有 int/bool/time.Duration
类型的配置。

配置列表参考官方文档：<https://pkg.go.dev/github.com/go-redis/redis#Options>

## 使用

框架通过`memdb.Get(name)`函数获取缓存实例，入参是配置名（去掉前缀），
返回的是`*redis.Client`对象。

```go
ctx, db := Get(ctx, "foo")
db.Set(ctx, "a", "123", 0)
```
