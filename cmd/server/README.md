# cmd/server

## 注册服务

自动注册服务请参考 [rpc/README.md](../../rpc/README.md#自动注册)。
注册外部服务请参考 `initMux` 方法，内部服务参考 `initInternalMux` 方法。

实现服务接口请参考 [server/README.md](../../server/README.md)。

## 启动服务

```bash
# 对外服务
go run main.go server --port=8080
# 对内服务
go run main.go server --port=8080 --internal
```
