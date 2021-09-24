# sqldb

sqldb 主要解决以下问题：

- 加载数据库配置
- 记录 sql 执行日志
- 上报 opentracing 追踪数据
- 汇总 prometheus 监控指标

核心思想是用`github.com/ngrok/sqlmw`把现有的`database/sql`驱动包起来，
拦截所有数据库操作进行观察。

## 配置

框架默认支持 sqlite 和 mysql。

每个数据库的配置需要指定一个名字，并添加`SQLDB_DSN_`前缀。

```yaml
# sqlite 配置示例
SQLDB_DSN_lite1 = "file:///tmp/foo.db"
# mysql 配置示例
SQLDB_DSN_mysql1 = "username:password@protocol(address)/dbname?param=value"
```

不同的驱动需要不同的配置内容：

- sqlite 请参考 <https://sqlite.org/c3ref/open.html>
- mysql 请参考 <https://github.com/mattn/go-sqlite3#connection-string>

## 使用

框架通过`sqldb.Get(name)`函数获取数据库实例，入参是配置名（去掉前缀），
返回的是`*sqlx.DB`对象。

框架会根据配置内容自动识别数据库驱动。

```go
ctx, db := sqldb.Get(ctx, "name")
db.ExecContext(ctx, "delete from ...")
```

## ORM

sqldb 提供简单的 Insert/Update/StructScan 方法，替换常用的 ORM 使用场景。

所有的 Model 都必须实现 Modler 接口，支持查询所属的表名和主键字段名。

比如我们定义一个 user 对象：

```go
type user struct {
	ID      int
	Name    string
	Age     int
	Created time.Time
}
func (u *user) TableName() string { return "users" }
func (u *user) KeyName() string   { return "id" }
```

保存对象：

```go
u := {Name:"foo", Age:18, Created:time.Now()}
result, err := db.Insert(&u)
```

更新对象：

```go
u.Name = "bar"
result, err := db.Update(&u)
```

查询对象：

```go
var u2 user
err := db.Get(&u2, "select * from users where id = ?", id)
```

## 现有问题

受限于 database/sql 驱动的设计，我们无法在提交或者回滚事务的时候确定总耗时。

目前只能监控 begin/commit/rollback 单个查询耗时，而非事务总耗时。


## 添加新驱动

如果想添加 sqlite 和 mysql 之外的数据库驱动（比如 postgres），需要做两件事。

首先是注册驱动，此部分代码需要添加到 observer.go 的`init()`函数：

```go
sql.Register("db-pg", sqlmw.Driver(pg.Driver{}, observer{}))
```

第一个参数是驱动的名字，可以随便取，但不能跟已经注册的冲突。建议添加`pb-`前缀。

然后就是根据配置内容指定驱动名字，这部分代码需要添加到 sqldb.go 的`Get()`函数：

```go
dsn := conf.Get("SQLDB_DSN_" + name)
var driver string
if strings.HasPrefix(dsn, "file://") {
	driver = "db-sqlite"
} else if strings.HasPrefix(dsn, "postgres://") {
	driver = "db-pq" // 这里需要跟注册的名字保持一致
} else {
	driver = "db-mysql"
}
```
