# db

mysql 基础库，开放有限接口，支持日志、opentracing 和 prometheus 监控。

# 配置

DB 配置，格式为 DB_${NAME}_DSN，内容参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name

必须设置 parseTime 选项。设置好了就可以通过 ${NAME} 获取 DB 连接池。

时区问题参考 https://www.jianshu.com/p/3f7fc9093db4

# 示例
```go
import "context"
import "sniper/util/db"

ctx := context.Background()
c := db.Get(ctx, "default")

sql := "insert into foo(id) values(1)"
q := SQLUpdate("foo", sql)
result, err := c.ExecContext(ctx, q)

// 执行 db 事务
err := c.ExecTx(ctx, func(ctx context.Context, tx db.Conn) error {
	sql := "insert into foo(id) values(1)"
	q := SQLUpdate("foo", sql)
	result, err := c.ExecContext(ctx, q)
	if err != nil {
		return err
	}

	sql := "insert into foo(id) values(2)"
	q := SQLUpdate("foo", sql)
	result, err = tx.ExecContext(ctx, q)
	if err != nil {
		return err
	}

	return nil
})
```
