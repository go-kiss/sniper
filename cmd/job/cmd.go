package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	httpd "net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"sniper/util"

	"sniper/util/conf"
	"sniper/util/ctxkit"
	"sniper/util/log"
	"sniper/util/metrics"
	"sniper/util/trace"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	crond "github.com/robfig/cron"
	"github.com/spf13/cobra"
)

type jobInfo struct {
	Name  string   `json:"name"`
	Spec  string   `json:"spec"`
	Tasks []string `json:"tasks"`
	job   func(ctx context.Context) error
}

func (j *jobInfo) Run() {
	j.job(context.Background())
}

var c = crond.New()

var jobs = map[string]*jobInfo{}
var httpJobs = map[string]*jobInfo{}

var port int

func init() {
	Cmd.Flags().IntVar(&port, "port", 8080, "metrics listen port")
}

// Cmd run job once or periodically
var Cmd = &cobra.Command{
	Use:   "job",
	Short: "Run job",
	Long: `You can list all jobs and run certain one once.
If you run job cmd WITHOUT any sub cmd, job will be sheduled like cron.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 不指定 handler 则会使用默认 handler
		server := &httpd.Server{Addr: fmt.Sprintf(":%d", port)}
		go func() {
			metricsHandler := promhttp.Handler()
			httpd.HandleFunc("/metrics", func(w httpd.ResponseWriter, r *httpd.Request) {
				util.GatherMetrics()

				metricsHandler.ServeHTTP(w, r)
			})

			httpd.HandleFunc("/ListTasks", func(w httpd.ResponseWriter, r *httpd.Request) {
				ctx := context.Background()
				span, ctx := opentracing.StartSpanFromContext(ctx, "ListTasks")
				defer span.Finish()

				w.Header().Set("x-trace-id", trace.GetTraceID(ctx))
				w.Header().Set("content-type", "application/json")

				buf, err := json.Marshal(httpJobs)
				if err != nil {
					w.WriteHeader(httpd.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				w.Write(buf)
			})

			httpd.HandleFunc("/RunTask", func(w httpd.ResponseWriter, r *httpd.Request) {
				ctx := context.Background()
				span, ctx := opentracing.StartSpanFromContext(ctx, "RunTask")
				defer span.Finish()

				w.Header().Set("x-trace-id", trace.GetTraceID(ctx))

				name := r.FormValue("name")
				job, ok := httpJobs[name]
				if !ok {
					w.WriteHeader(httpd.StatusNotFound)
					w.Write([]byte("job " + name + " not found\n"))
					return
				}

				if err := job.job(ctx); err != nil {
					w.WriteHeader(httpd.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("%+v", err)))
					return
				}

				w.Write([]byte("run job " + name + " done\n"))
			})

			httpd.HandleFunc("/monitor/ping", func(w httpd.ResponseWriter, r *httpd.Request) {
				w.Write([]byte("pong"))
			})

			if err := server.ListenAndServe(); err != nil {
				panic(err)
			}
		}()

		go func() {
			conf.OnConfigChange(func() { util.Reset() })
			conf.WatchConfig()

			c.Run()
		}()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
		<-stop

		var wg sync.WaitGroup
		go func() {
			wg.Add(1)
			defer wg.Done()

			c.Stop()
		}()
		go func() {
			wg.Add(1)
			defer wg.Done()

			err := server.Shutdown(context.Background())
			if err != nil {
				panic(err)
			}
		}()
		wg.Wait()
	},
}

var cmdList = &cobra.Command{
	Use:   "list",
	Short: "List all jobs",
	Long:  `List all jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		for k, v := range jobs {
			fmt.Printf("%s [%s]\n", k, v.Spec)
		}
		for k, v := range httpJobs {
			fmt.Printf("%s [%s]\n", k, v.Spec)
		}
	},
}

// once 命令参数，可以在 cron 中使用
// sniper job once foo bar 则 onceArgs = []string{"bar"}
// sniper job once foo 1 2 3 则 onceArgs = []string{"1", "2", "3"}
var onceArgs []string
var onceHttpJob bool

var cmdOnce = &cobra.Command{
	Use:   "once job",
	Short: "Run job once",
	Long:  `Run job once.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		onceArgs = args[1:]
		var job *jobInfo
		if onceHttpJob {
			job = httpJobs[name]
		} else {
			job = jobs[name]
		}
		if job != nil {
			job.job(context.Background())
		}
	},
}

func init() {
	cmdOnce.Flags().BoolVarP(&onceHttpJob, "http", "", false, "运行 http 任务")
	Cmd.AddCommand(
		cmdList,
		cmdOnce,
	)
}

// http 注册的任务需要 http 触发
// spec 采用 unix crontab 语法，不支持秒!!!
func http(name string, spec string, job func(ctx context.Context) error, args ...string) {
	if _, ok := httpJobs[name]; ok {
		panic(name + "is used")
	}

	if spec == "@manual" {
		return
	}

	schedule := "@once" // 只触发一次
	if strings.HasPrefix(spec, "@") {
		switch {
		case strings.Contains(spec, "every"):
			// TODO scheduler trans
		default:
			schedule = spec // @hourly @daily ...
		}
	} else {
		schedule = spec
	}

	httpJobs[name] = regjob(name, schedule, job, args)
	return
}

// sepc 参数请参考 https://godoc.org/github.com/robfig/cron
func cron(name string, spec string, job func(ctx context.Context) error) {
	if _, ok := jobs[name]; ok {
		panic(name + " is used")
	}

	j := regjob(name, spec, job, []string{})
	jobs[name] = j

	if spec == "@manual" {
		return
	}

	if err := c.AddJob(spec, j); err != nil {
		panic(err)
	}
}

func manual(name string, job func(ctx context.Context) error) {
	cron(name, "@manual", job)
}

func regjob(name string, spec string, job func(ctx context.Context) error, tasks []string) (ji *jobInfo) {
	j := func(ctx context.Context) (err error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "Cron")
		defer span.Finish()

		span.SetTag("name", name)
		ctx = ctxkit.WithTraceID(ctx, trace.GetTraceID(ctx))

		logger := log.Get(ctx)

		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprintf("%+v stack: %s", r, string(debug.Stack())))
				logger.Error(err)
			}
		}()

		if conf.GetBool("JOB_PAUSE") {
			logger.Errorf("skip cron job %s[%s]", name, spec)
			return
		}

		code := "0"
		t := time.Now()
		if err = job(ctx); err != nil {
			logger.Errorf("cron job error: %+v", err)
			code = "1"
		}
		d := time.Since(t)

		metrics.JobTotal.WithLabelValues(code).Inc()

		logger.WithField("cost", d.Seconds()).Infof("cron job %s[%s]", name, spec)
		return
	}

	ji = &jobInfo{Name: name, Spec: spec, job: j, Tasks: tasks}
	return
}

// 已废弃，请使用 cron 或 manual
func addJob(name string, spec string, job func(ctx context.Context) error) {
	cron(name, spec, job)
}
