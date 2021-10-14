# conf

默认从 sniper.toml 加载配置。虽然 sniper.toml 支持复杂的数据结构，
但框架要求只能设置 k-v 型配置。目的是为了跟环境变量相兼容。

sniper.toml 中的所有配置项都可以使用环境变量覆写。

如果配置文件不在项目根目录，则可以通过环境变量 `CONF_PATH` 指定。

框架还会自动监听 sniper.toml 内容变更，发现变更会自动热加载。

最后，配置跟环境变量一样，不区分大小写字母。

# 示例
```go
import "github.com/go-kiss/sniper/pkg/conf"

b := conf.GetBool("IS_SHUTTING_DOWN")
```
