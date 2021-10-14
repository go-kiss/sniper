# rpc

接口定义层，基于 protobuf 严格定义 RPC 接口路由、参数和文档。

## 目录结构

通常一个服务一个文件夹。服务下有版本，一个版本一个文件夹。内部服务一般使用 `v0` 作为版本。

一个版本可以定义多个 service，每个 service 一个 proto 文件。

典型的目录结构如下：
```
rpc/user # 业务服务
└── v0   # 服务版本
    ├── echo.go        # rpc 方法实现，方法签名由脚手架自动生成
    ├── echo.pb.go     # protobuf message 定义代码[自动生成]
    ├── echo.proto     # protobuf 描述文件[业务方定义]
    └── echo.twirp.go  # rpc 接口和路由代码[自动生成]
```

## 定义接口

服务接口使用 [protobuf](https://developers.google.com/protocol-buffers/docs/proto3#services) 描述。
```proto
syntax = "proto3";

package user.v0; // 包名，与目录保持一致

// 服务名，只要能定义一个 service
service Echo {
  // 服务方法，按需定义
  rpc Hello(HelloRequest) returns (HelloResponse);
}

// 入参定义
message HelloRequest {
  // 字段定义，如果使用 form 表单传输，则只支持
  // int32, int64, uint32, unint64, double, float, bool, string,
  // 以及对应的 repeated 类型，不支持 map 和 message 类型！
  // 框架会自动解析并转换参数类型
  // 如果用 json 或 protobuf 传输则没有限制
  string message = 1; // 这是行尾注释，业务方一般不要使用
  int32 age = 2;
  // form 表单格式只能部分支持 repeated 语义
  // 但客户端需要发送英文逗号分割的字符串
  // 如 ids=1,2,3 将会解析为 []int32{1,2,3}
  repeated int32 ids = 3;
}

message HelloMessage {
  string message = 1;
}

// 出参定义,
// 理论上可以输出任意消息
// 但我们的业务要求只能包含 code, msg, data 三个字段，
// 其中 data 需要定义成 message
// 开源版本可以怱略这一约定
message HelloResponse {
  // 业务错误码[机读]，必须大于零
  // 小于零的主站框架在用，注意避让。
  int32 code = 1;
  // 业务错误信息[人读]
  string msg = 2;
  // 业务数据对象
  HelloMessage data = 3;
}
```

### GET 请求

有些业务场景需提供 GET 接口，原生的 twirp 框架并不支持。但 sniper 框架是支持的。

只需要在 `hook.RequestReceived` 阶段调用 `ctx = twirp.WithAllowGET(ctx, true)` 将 GET 开关注入 ctx 即可。

但原则上不建议使用 GET 请求。

### 文件下载

有些业务场景需提供 json/protobuf 之外的数据，如 xml、txt 甚至是 xlsx。

sniper 为这类情况留有「后门」。只需要定义并返回一个特殊的 response 消息：
```proto
// 消息名可以随便取
message DownloadMsg {
    // content_type 内容用于设置 http 的 content-type 字段
    string content_type = 1;
    // data 内容会直接以 http body 的形式发送给调用方
    bytes data = 2;
}
```

## 接口映射

- 请求方法 **POST**
- 请求路径 **/twirp**/package.Service/Method
- 请求协议 http/1.1、http/2
- Content-Type
  - application/x-www-form-urlencoded
  - application/json
  - application/protobuf
- 请求内容
  - urlencoded 字符串
  - json
  - protobuf

最新版的[protobuf-gen-twirp](./cmd/protoc-gen-twirp)生成的 `*.twirp.go` 文件已经
不再硬编码 `/twirp` 前缀。接口前缀可以通过 `RPC_PREFIX` 配置项控制，默认前缀为 `/api`。

表单请求
```
POST /user.v0.Echo/Hello HTTP/1.1
Host: example.com
Content-Type: application/x-www-form-urlencoded
Content-Length: 19

message=hello&age=1

HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 27

{"message":"Hello, World!"}
```
json 请求
```
POST /user.v0.Echo/Hello HTTP/1.1
Host: example.com
Content-Type: application/json
Content-Length: 19

{"message":"hello","age":1}

HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 27

{"message":"Hello, World!"}
```

原始英文协议在[这里](../util/twirp/PROTOCOL.md)

## 生成代码

```bash
# 首次使用需要安装 protoc-gen-twirp 工具
make cmd
# 针对指定服务
protoc --go_out=. --twirp_out=. echo.proto

# 针对所有服务
find rpc -name '*.proto' -exec protoc --twirp_out=. --go_out=. {} \;

# 建议直接使用框架提供的 make 规则
make rpc
```

生成的文件中 `*.pb.go` 是由 protobuf 消息的定义代码，同时支持 protobuf 和 json。`*.twirp.go` 则是 rpc 路由相关代码。

## 自动注册

sniper 提供的脚手架可以自动生成 proto 模版、server 模版，并注册路由。
运行以下命令：
```bash
go run cmd/sniper/main.go rpc --server=foo --service=echo
```
会自动生成：
```
rpc
└── foo
    └── v1
        ├── echo.go
        ├── echo.pb.go
        ├── echo.proto
        └── echo.twirp.go
```

## 实现接口

服务接口定义在 rpc 目录对应的 echo.twirp.go 中，是自动生成的。

接口实现代码则会自动生成并保存到 echo.go 中。

```go
package foo_v0

import (
	// 标准库单列一组
	"context"

	// 框架库单列一组
	"sniper/dao/login"
	"sniper/pkg/conf"
)

// 服务对象，约定为 Server
type EchoServer struct{}

// 接口实现，三步走：处理入参、调用服务、返回出参
func (s *EchoServer) ClearLoginCache(ctx context.Context, req *ClearRequest) (*EmptyReply, error) {
	// 处理入参
	mid := req.GetMid()

	// 调用 service 层或者 dao 层完成业务逻辑
	login.ClearUID(ctx, mid)

	// 返回出参
	reply := &EmptyReply{}

	return reply, nil
}
```

## 注册服务

请参考 [cmd/server/README.md](../cmd/server/README.md)。

## 错误处理

### 异常/错误

**错误** 是 __计划内__ 的情形，例如用户输入密码不匹配、用户余额不足等等。
**异常** 是 __计划外__ 的情形，例如用户提交的参数类型跟接口定义不匹配、DB 连接超时等等。

**错误** 可以认为是一种特殊的“正常情况”, **异常** 则是真正的“不正常情况”。

### 处理错误

客户端需要根据不同业务需求处理 **错误**, 例如用户未登录则需要跳转到登录页面。所以，我需要使用错误码来返回错误信息。

处理代码示例如下：
```go
resp := &pb.Resp{}

resp.Code = 100
resp.Msg = "Need Login"

return nil, resp
```
以上代码会返回如下 HTTP 信息：
```
HTTP/1.1 200 OK
Content-Length: 355
Content-Type: application/json
Date: Tue, 14 Aug 2018 03:05:41 GMT
X-Trace-Id: 3kclnknyzmamo

{
    "code": 100,
    "msg": "Need Login",
    "data": {}
}
```

### 处理异常

正常的客户端会严格按照接口定义调用接口，只有客户端有 bug 或者服务端有问题的时候才会遇到 **异常**。
在这种情况下，首先，我们无法从错误中恢复；其次，这类错误的处理方式跟具体的业务没有关系的；最后，我们需要 **及时发现** 这类问题并修复。
所以，我们需要使用 HTTP 的 4xx 和 5xx 状态码来返回错误信息。

处理代码示例如下：
```go
import "sniper/pkg/errors"
// ...

// 这是客户端问题，返回 HTTP 4xx 状态码
if req.ID <= 0 {
	return nil, errors.InvalidArgumentError("id", "must > 0")
}

// HTTP/1.1 400 Bad Request
// Content-Length: 104
// Content-Type: application/json
// Date: Tue, 14 Aug 2018 03:09:30 GMT
// X-Trace-Id: kg1od386gjto
//
// {
//     "code": "invalid_argument",
//     "meta": {
//         "argument": "page_size"
//     },
//     "msg": "page_size page_size must be > 0"
// }

// 这是服务端问题，返回 HTTP 5xx 状态码
if err := bookshelf.AddFavorite(ctx, id); err != nil {
	return nil, err
}

// HTTP/1.1 500 Internal Server Error
// Content-Length: 112
// Content-Type: application/json
// Date: Wed, 15 Aug 2018 08:50:47 GMT
// X-Trace-Id: 3njq5120j3c1n
//
// {
//     "code": "internal",
//     "meta": {
//         "cause": "*net.OpError"
//     },
//     "msg": "dial tcp :0: connect: can't assign requested address"
// }
```

我们可以通过 SLB 报警及时发现此类错误并减少业务损失。
