# trace

框架支持 [opentracing](https://opentracing.io/)，默认集成 [jaeger](https://github.com/jaegertracing/jaeger-client-go)。

如果想开启 jaeger 收集 opentracing 数据，需要以下配置：

- `JAEGER_AGENT_HOST` jaeger 服务器IP或域名，默认为 127.0.0.1
- `JAEGER_AGENT_PORT` jaeger 服务器端口，默认为 6831
- `JAEGER_SAMPLER_PARAM` 采样率，0-1 之间的浮点数，默认为0，也就是不采集

个人开发环境可以使用 docker 体验：

```bash
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.25
```

启动后访问 <http://127.0.0.1:16686> 即可打开查询界面。
