# service

业务逻辑层，处于 rpc 层和 dao 层之间。service 只能通过 dao 层获取数据。

业务接口必须接受 `ctx context.Context` 对象，并向下传递。

## 错误日志
