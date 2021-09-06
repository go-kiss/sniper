# Sniper['snaɪpər] 轻量级业务框架

[Sniper 的前世今生](./thought.md)

有兴趣的同学也可以加微信 `taoshu-in` 讨论，拉你进群。

## 系统要求

1. 类 UNIX 系统
2. go v1.12+
3. [protoc](https://github.com/google/protobuf)
4. [protoc-gen-go](https://github.com/golang/protobuf/tree/master/protoc-gen-go)

## 目录结构

```
├── cmd         # 服务子命令
├── dao         # 数据访问层
├── main.go     # 项目总入口
├── rpc         # 接口描述文件
├── server      # 控制器层
├── service     # 业务逻辑层
├── sniper.toml # 配置文件
└── util        # 业务工具库
```

## 快速入门

- [定义接口](./rpc/README.md)
- [实现接口](./rpc/README.md)
- [注册服务](./cmd/server/README.md)
- [启动服务](./cmd/server/README.md)
- [配置文件](./util/conf/README.md)
- [日志系统](./util/log/README.md)
- [指标监控](./util/metrics/README.md)
- [链路追踪](./util/trace/README.md)
