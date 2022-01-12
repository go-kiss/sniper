# Sniper 轻量级业务框架

Sniper 是一套轻量级但又很现代化的业务框架。轻量体现在只集成了最必要的功能，现代
则体现在接口描述IDL、可观测、强大的脚手架等方面。

Sniper 框架从 2018 年开发并开源，在我们业务生产环境平稳运行，至少可以应对五百万 
DAU量级的业务。我们也不断把多年的生产实践经验固化到 Sniper 框架，希望能帮助更多 
的朋友。

有兴趣的同学也可以加我的微信`taoshu-in`我拉大家进群讨论。

## 系统要求

Sniper 仅支持 UNIX 环境。Windows 用户需要在 WSL 下使用。

环境准备好之后，需要安装以下工具的最新版本：

- go
- git
- make
- [protoc](https://github.com/google/protobuf)

## 快速入门

安装 sniper 脚手架：

```bash
go install github.com/go-kiss/sniper/cmd/sniper@latest
```

创建一个新项目：

```bash
sniper new --pkg helloworld
```

切换到 helloworld 目录。

运行服务：

```bash
CONF_PATH=`pwd` go run main.go http
```

使用 [httpie](https://httpie.io) 调用示例接口：

```bash
http :8080/api/foo.v1.Bar/Echo msg=hello
```

应该会收到如下响应内容：

```
HTTP/1.1 200 OK
Content-Length: 15
Content-Type: application/json
Date: Thu, 14 Oct 2021 09:49:16 GMT
X-Trace-Id: 08c408b0a4cd12c0

{
    "msg": "hello"
}
```

## 深入理解

Sniper 框架几乎每一个目录下都有 README.md 文件，建议仔细阅读。

如需了解 Sniper 框架的工作原理和设计原则，请移步我的[博客](https://taoshu.in/go/sniper.html)。
