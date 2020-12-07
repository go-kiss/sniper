# Sniper 的前世今生

Sniper 起源于一项新业务。在转岗之前，我一直在 L 部门写 PHP 代码，遇到过如下问题：

- 基于 TCP 的 RPC 协议，我们都称之为 **Weisai-RPC**
- 手工维护 RPC 文档，难以及时更新
- 手写代码处理 RPC 入参，难以保证参数类型，如数字 `1` 和字符串 `"1"` 的区别
- 无法方便地查询一个请求对应的所有日志
- 服务拆分得很细，难以进行调用链路追踪
- 使用 JSON 做为配置，难改难认
- 难以监控服务运行状态
- 代码分层标准不统一
- 没有单元测试

大约在 2018 年的六月底，我得知要去新的 C 部门做新业务。没有任何历史包袱，我马上着手准备，希望能全方位的解决上面提到的问题。

## Go 语言

首先要解决语言选择的问题。PHP 是最熟悉的，但从过去的经验来看，无论从性能还是从代码可维护性方面考虑，PHP 都不是一个好的选择。当时有两种选择，一个是 Java，另一个是 go。平心而论，Java 是要比 Go 要成熟得多。但 Go 更加简单轻便，从 PHP 过渡成本更低。而且当时公司正在推动用 Go 重写原有的 Java 项目。自然就选了 Go。

## RPC 协议
有了语言，接下来就要确定通信协议。首先不要使用 REST 风格接口。 REST 中看不中用。REST 的核心是资源和状态，所有的变更都对应状态的转变。

对于简单的场景，REST 看似完美，如：`GET /user/123` 表示查询。

但如果是发送一条短信呢？一种方案是使用 `POST /sms` 表示创建一条**短信资源**，另一种方案则是 `POST /sms:send` 直接发送。

但不管哪种方式，都不如 RPC 调用直观，其原因有二：
- 一是 http 的方法（GET, POST, PUT, DELETE 等）太少，基本都是面向静态资源的，表达能力有限
- 二是将业务过程转成资源状态变化本身就比较烧脑，而且存在无法转化的场景

REST 还有一个比较大的问题就是 url 中有数字 id，统计 prometheus 监控指标的时候必须做**归一化**处理。

所以，不用 REST。

## Weisai RPC

这得从原来在 L 部门用的 Weisai-RPC 说起。该 RPC 基于 TCP 传输，消息结构如下：
```c
typedef struct swoole_message {
    uint32_t header_magic;     // magic 字段 默认2233
    uint32_t header_ts;        // unix时间戳
    uint32_t header_check_sum; // 校验和, 暂未定义, 默认为0
    uint32_t header_version;   // 版本号
    uint32_t header_reserved;  // 保留字段, 默认0, live-api转发时设置为1
    uint32_t header_seq;       // 序列号
    uint32_t header_len;       // body长度
    char cmd[32];              // 命令字符串
                               // 格式 {message_type}controller.method,
                               // message_type 0 request, 1 response
                               // 长度没满右端补充\0, 超过自动右端截断.
    char* body;                // 可变 长度为header_len 格式为JSON:
                               // {"header":..., "body":....}
} rpc_message_t;
```
典型的面向 c 语言的设计，方便 c 语言解析，但不太灵活。

比如，cmd 字段只有 32 字节，也就是说接口名字最多只能是 32 字节。还有 body 是字符串，但实际传输的是 JSON，需要二次解析。使用结构化二进制消息就是为了提高解析速度，但这种改进跟 JSON 解码相比又可以忽略。所以，这种混合型的设计除了看上去比较复杂以外，确实没什么优点了。

因为没有采用 HTTP 协议，后来不得不在 body 中定义了 header 字段用来传输 HTTP 请求的 header。像 nginx, curl, tcpdump 这样的标准也基本上无法正常使用。为此，还专门引入了一个接入层负责 RPC 和 HTTP 之间的相互转换。

切实体会到了 Weisai-RPC 的不便之后，我决定业务 RPC 协议只用 HTTP 传输，原则上不使用二进制消息格式。

## 关于 gRPC

说到 HTTP 就不得不说说 gRPC。gRPC 是 Google 开放的一种 RPC 协议，其主要特性：
- 只支持 protobuf 编码
- 强依赖 HTTP2 协议
- 支持 stream 接口
- 每个消息都有五字节的二进制前缀
其他细节请参考 [PROTOCOL-HTTP2](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md)。

protobuf 本身是支持 JSON 的，不明白为什么 gRPC 的实现不支持。而支持 stream 接口则是 gRPC 的一大特色，使 gRPC 能够胜任诸如语音实时识别等场景。但这一类场景是比较少见的。我们绝大多数业务场景都是一问一答的。为了实现这个 stream 特性，gRPC 不得不依赖 HTTP2，不得不自行定义了一种有固定五字节头的消息格式。与此同时，gRPC 也就放弃了 HTTP 协议原生的压缩功能，也没法使用 HTTP 协议的 content-length 头传递消息长度。这也是 gRCP 消息五字节头的功能所在，头一个字节表示是否压缩，后四个字节表示消息长度。

有个所谓的 **2-8** 原则：
> 一般只用 **20%** 的代码就可以解决 **80%** 的问题。但要想解决剩下 **20%** 的问题的话，则需要额外 **80%** 的代码。

gRPC 的 stream 接口就是剩下的 **20% 的问题**。

gRPC 还有个 web 支持的问题。浏览器的 js 无法使用 HTTP2 的特性，所以不能直接与 gRPC 服务通信。于是有了 [grpc-web](https://github.com/grpc/grpc-web)，还有 [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)。

所以，如果没有 stream 接口需求，则完全没有必要使用 gRPC；如果真的有这类需求，也不可能太多，直接使用原生 TCP/WebSocket 协议开发也不是难事。

最终我们选择了 [twirp](https://github.com/twitchtv/twirp)。twirp 可以看作是简化版的 gRPC，同样用 protobuf 描述，不依赖 HTTP2，同时支持 protobuf 和 JSON，没有五字节的二进制前缀。但我们对原生的 twirp 做了修改，形成了自己的[版本](https://github.com/bilibili/twirp)，主要改动就是添加了对 **www-form-urlencoded** 编码格式的支持，这是移动端的历史包袱导致的，没办法。

现在的移动端使用 www-form-urlencoded 编码，更加简单；管理后台使用 JSON 编码，更加灵活。如果对性能有要求也可以使用 protobuf 编码，但没目前没有用，估计也不会有人喜欢用。

## 接口文档
使用 proto 描述 RPC 接口有一个问题，就是接口说明分了 request, response 和 service，比较分散，尤其是要用到嵌套 message 的时候，对移动端开发同学很不友好。目前也一些文档生成工具，比如：[protoc-gen-doc](https://github.com/lvht/protoc-gen-markdown/blob/master/hello.md)。但 protoc-gen-doc 也是为不同 message 生成对应文档，使用者需要在文档的不同部分来回跳转，很不直观。所以我们开发了 [protoc-gen-markdown](https://github.com/lvht/protoc-gen-markdown)。这是生成的[文档示例](https://github.com/lvht/protoc-gen-markdown/blob/master/hello.md)。最终，我们给 gitlab 加了一个 webhook，当有新分支创建或者更新的时候会自动生成 markdown 文档并进而转化成 html 文档，彻底解决了文档同步的问题。

protoc-gen-markdown 也不完美。它无法正确处理 proto 中的 map 消息。但我们在业务中没有用到这种类型，所以没有受到影响。但这始终是个问题。protoc-gen-markdown 最早是跟 twirp 的改造一起进行的。最早的提交记录是从 2018 年 7 月 3 日开始的，主要功能到 7 月 7 日就完成了，到现在也没有大的变动。

## 配置系统
解决了通信问题之后，接下来要设计配置系统。

在 L 部门的时候都是用 JSON 做配置。JSON 一方面对格式要求比较高，比如列表最后一个元素之后不能加逗号等；另一方面不支持注释，时间长了很难弄清各配置项的含义。还有就是 JSON 很灵活，导致很多业务配置层层嵌套，不好读、不敢改。

鉴于之前的经验，我们放弃了 JSON，最终选择了 [toml](https://github.com/toml-lang/toml)。而且框加要求所有配置只能是 k-v 型字符串的。如果业务代码要用复杂的配置，则需要自行处理反序列化逻辑。因为是 k-v 型的，所以很容易兼容**环境变量**，所有的配置项都可以通过环境变量覆盖。最后就是框架支持配置的热更新，会实时读取配置文件内容的变更。

我们也没有重复造轮子，配置的解析和加载都是通过 [viper](https://github.com/spf13/viper) 完成的。

## 日志与监控
日志组件选用 [logrus](https://github.com/sirupsen/logrus)。没别的原因，就是 star 比较多。logrus 支持不同的 formatter，开发环境会将日志写到标准输出设备，其他环境会通过 lancer 写到 elk（这一部分不适合开源）。

框架在处理请求的时候会创建一个 opentracing 的 span。这个 span 是有一个 trace-id 的。框架会把这个 trace-id 注入到 ctx 中。我们希望相关的日志都要带有这个 trace-id，所以需要通过 `sniper/util/log.Get(ctx context.Context)` 方法来获取 logger 实例，使用获取的实例记录日志会自动输出 trace-id。框架在输出响应内容的时候也会自动在 header 中加上这个 trace-id。

公司内部有个叫 dapper 组件，但没有 opentracing sdk。框架自己提供了一个，但这一部分不适合开源。

好在是适配了 opentracing，大家可以很方便的集成 jaeger 等组件。

## 基础组件
主要的基础组件有三个，分别是 HTTP 客户端、mysql 客户端、[memcache 客户端](https://github.com/bilibili/memcache)。[redis 客户端](https://github.com/bilibili/redis)是后来加入的，现在还没在业务中使用。

Sniper 对基础组件提供统一封装，主要解决以下问题：
- 加载配置
- 处理 ctx
- 输出日志
- 支持 opentracing
- 统计 prometheus 指标

现在很少有框架会注意到这些方面，尤其是后三条。大家关注更多的往往是性能，往往是框架代码是否优雅。估计只有在生产环境摸爬滚打过几次才会对这些东西产生共鸣。

### 关于 ORM
很多框架都提供 ORM 组件，但 sniper 不然。不推荐使用 ORM，原因如下：
- ORM 固然方便，但会隐藏 SQL 查询细节，不利于程序员全盘掌握 db 查询情况。
- ORM 用法并不统一，相对 SQL 标准有额外的学习负担。
- ORM 无法覆盖所有有 SQL 查询，在特定业务场景下仍需要写原生 SQL。
- ORM 大多基于反射，有一定的性能损失。
- 业务代码一般会有数据访问层（DAO），即便引入 ORM，也只局限在 DAO 层。

### 关于集群
Sniper 框架的 memcache 和 redis 组件都不支持集群的，而且是有意不支持甚至是将已有的相关代码直接删除。

为什么呢？我们认为这些细节不应该是一个业务框架要关心的内容。这些内容应该交给统一的中间件处理。业务代码连中间件，根本无需感知集群的存在。对于 memcache，我们生产环境用的是 [twemproxy](https://github.com/twitter/twemproxy)，对于 redis 和 http 服务，我们用的是 [envoy](https://www.envoyproxy.io/)。

我们坚信，未来一定是 service-mesh 的世界，诸如服务发现、负载均衡、限流熔断这一类的功能应该交由 mesh 服务处理。让我们拭目以待。

## 单元测试
单元测试部分不适合开源，只能分享一些相关的思考。

没有单元测试，就很难有真正的积累。我们的核心业务逻辑基本都有单元测试覆盖。有一次要改支付逻辑，我改完跑通测试后直接移交测试，测试通过，直接上线，一气呵成。我甚至都没自己用 curl 调一下接口，因为我知道，单元测试已经覆盖的已知的关键流程。

这当然不是什么值得炫耀的事情。但有效的单元测试确实对提高代码的质量有很大的裨益。

但怎么测才好呢？关键在 mock。Go 对 mock 并不是很友好。而且如果 mock 多了，一方面会极大降低写测试用例的体验；另一方面会导致测试用例真就成单元测试了，可能出现各单元都没问题，但整个系统有问题的情况。

所以，写测试一定要简单，测试逻辑一定要有效。为实现这两个目标，我们定了两条规则：
- 外部 http 请求一律 mock，这个基于 [jarcoal/httpmock](https://github.com/jarcoal/httpmock)
- mysql, memcache, redis 直接起服务，各测试用例自行维护自己的测试数据集

为了进一步降低编写测试用例的复杂度，我们还提供了自动同步表结构和导入种数据的功能。如果测试用例不想手工维护测试数据集，则可以将相关数据写种子数据集。测试框架会自动导入。

## 总结
引入 sniper 框架快一年了，基本上解决了在 L 部门遇到的问题，无论在线下开发、联调和测试效率方面，还是线上运行、排错效率方面，都有不俗的表现。
