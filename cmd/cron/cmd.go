package cron

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"sniper/pkg"
	"sniper/pkg/conf"
	"sniper/pkg/log"
	"sniper/pkg/trace"

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
	Use:   "cron",
	Short: "Run cron job",
	Long: `You can list all jobs and run certain one once.
If you run job cmd WITHOUT any sub cmd, job will be sheduled like cron.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 不指定 handler 则会使用默认 handler
		server := &http.Server{Addr: fmt.Sprintf(":%d", port)}
		go func() {
			http.HandleFunc("/metrics", promhttp.Handler())

			http.HandleFunc("/ListTasks", func(w http.ResponseWriter, r *http.Request) {
				ctx := context.Background()
				span, ctx := opentracing.StartSpanFromContext(ctx, "ListTasks")
				defer span.Finish()

				w.Header().Set("x-trace-id", trace.GetTraceID(ctx))
				w.Header().Set("content-type", "application/json")

				buf, err := json.Marshal(httpJobs)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				w.Write(buf)
			})

			http.HandleFunc("/RunTask", func(w http.ResponseWriter, r *http.Request) {
				ctx := context.Background()
				span, ctx := opentracing.StartSpanFromContext(ctx, "RunTask")
				defer span.Finish()

				w.Header().Set("x-trace-id", trace.GetTraceID(ctx))

				name := r.FormValue("name")
				job, ok := httpJobs[name]
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("job " + name + " not found\n"))
					return
				}

				if err := job.job(ctx); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("%+v", err)))
					return
				}

				w.Write([]byte("run job " + name + " done\n"))
			})

			http.HandleFunc("/monitor/ping", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("pong"))
			})

			if err := server.ListenAndServe(); err != nil {
				panic(err)
			}
		}()

		go func() {
			conf.OnConfigChange(func() { pkg.Reset() })
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
	Short: "List all cron jobs",
	Long:  `List all cron jobs.`,
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
// sniper cron once foo bar 则 onceArgs = []string{"bar"}
// sniper cron once foo 1 2 3 则 onceArgs = []string{"1", "2", "3"}
var onceArgs []string

var cmdOnce = &cobra.Command{
	Use:   "once job",
	Short: "Run job once",
	Long:  `Run job once.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		onceArgs = args[1:]
		job := jobs[name]
		if job != nil {
			job.job(context.Background())
		}
	},
}

func init() {
	Cmd.AddCommand(
		cmdList,
		cmdOnce,
	)
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

		t := time.Now()
		if err = job(ctx); err != nil {
			logger.Errorf("cron job error: %+v", err)
		}
		d := time.Since(t)

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
