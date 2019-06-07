# Sniper['snaɪpər] 轻量级业务框架

[Sniper 的前世今生](./thought.md)

## 系统要求

1. 类 UNIX 系统
2. go v1.12+
3. [protoc](https://github.com/google/protobuf)
4. [protoc-gen-go](https://github.com/golang/protobuf/tree/master/protoc-gen-go)
5. [protoc-gen-twirp](https://github.com/bilibili/twirp/tree/master/protoc-gen-twirp)

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
- [实现接口](./server/README.md)
- [注册服务](./cmd/server/README.md)
- [启动服务](./cmd/server/README.md)
- [配置文件](./util/conf/README.md)
- [日志系统](./util/log/README.md)
- [数据库](./util/db/README.md)
- [memcache](./util/mc/README.md)
- [redis](./util/redis/README.md)
- [metrics](./util/metrics/README.md)
- [tracing](./util/trace/README.md)
