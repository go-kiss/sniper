package http

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"sniper/pkg"
	"sniper/pkg/conf"
	"sniper/pkg/log"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var port int

// Cmd run http server
var Cmd = &cobra.Command{
	Use:   "http",
	Short: "Run http server",
	Long:  `Run http server`,
	Run: func(cmd *cobra.Command, args []string) {
		main()
	},
}

func init() {
	Cmd.Flags().IntVar(&port, "port", 8080, "listen port")
}

var server *http.Server

type panicHandler struct {
	handler http.Handler
}

// 从 http 标准库搬来的
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

var logger = log.Get(context.Background())

func startSpan(r *http.Request) (*http.Request, opentracing.Span) {
	operation := "ServerHTTP"

	ctx := r.Context()
	var span opentracing.Span

	tracer := opentracing.GlobalTracer()
	carrier := opentracing.HTTPHeadersCarrier(r.Header)

	if spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, carrier); err == nil {
		span = opentracing.StartSpan(operation, ext.RPCServerOption(spanCtx))
		ctx = opentracing.ContextWithSpan(ctx, span)
	} else {
		span, ctx = opentracing.StartSpanFromContext(ctx, operation)
	}

	ext.SpanKindRPCServer.Set(span)
	span.SetTag(string(ext.HTTPUrl), r.URL.Path)

	return r.WithContext(ctx), span
}

func (s panicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r, span := startSpan(r)

	defer func() {
		if rec := recover(); rec != nil {
			ctx := r.Context()
			log.Get(ctx).Error(rec, string(debug.Stack()))
		}
		span.Finish()
	}()

	origin := r.Header.Get("Origin")
	suffix := conf.Get("CORS_ORIGIN_SUFFIX")

	if origin != "" && suffix != "" && strings.HasSuffix(origin, suffix) {
		w.Header().Add("Access-Control-Allow-Origin", origin)
		w.Header().Add("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Origin,No-Cache,X-Requested-With,If-Modified-Since,Pragma,Last-Modified,Cache-Control,Expires,Content-Type,Access-Control-Allow-Credentials,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Cache-Webcdn,Content-Length")
	}

	if r.Method == http.MethodOptions {
		return
	}

	s.handler.ServeHTTP(w, r)
}

func main() {
	reload := make(chan int, 1)
	stop := make(chan os.Signal, 1)

	conf.OnConfigChange(func() { reload <- 1 })
	conf.WatchConfig()
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	startServer()

	for {
		select {
		case <-reload:
			pkg.Reset()
		case sg := <-stop:
			stopServer()
			// 仿 nginx 使用 HUP 信号重载配置
			if sg == syscall.SIGHUP {
				startServer()
			} else {
				pkg.Stop()
				return
			}
		}
	}
}

func startServer() {
	logger.Info("start server")

	rand.Seed(int64(time.Now().Nanosecond()))

	mux := http.NewServeMux()

	initMux(mux)

	var handler http.Handler

	handler = panicHandler{handler: mux}

	if prefix := conf.Get("RPC_PREFIX"); prefix != "" && prefix != "/" {
		handler = http.StripPrefix(prefix, handler)
	}

	http.Handle("/", handler)
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/monitor/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	addr := fmt.Sprintf(":%d", port)
	server = &http.Server{
		IdleTimeout: 60 * time.Second,
	}

	// 配置下发可能会多次触发重启，必须等待 Listen() 调用成功
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		// 本段代码基本搬自 http 标准库
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		wg.Done()

		err = server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
		if err != http.ErrServerClosed {
			panic(err)
		}
	}()

	wg.Wait()
}

func stopServer() {
	logger.Info("stop server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal(err)
	}

	pkg.Reset()
}
