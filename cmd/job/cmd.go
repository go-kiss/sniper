package job

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
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
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
)

type jobInfo struct {
	spec string
	job  func()
}

var c = cron.New()
var jobs = map[string]jobInfo{}

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
		go func() {
			addr := fmt.Sprintf(":%d", port)
			err := http.ListenAndServe(addr, promhttp.Handler())
			if err != nil {
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

		c.Stop()
	},
}

var cmdList = &cobra.Command{
	Use:   "list",
	Short: "List all jobs",
	Long:  `List all jobs.`,
	Run: func(cmd *cobra.Command, args []string) {
		for k, v := range jobs {
			fmt.Printf("%s [%s]\n", k, v.spec)
		}
	},
}

// once 命令参数，可以在 cron 中使用
// sniper job once foo bar 则 onceArgs = []string{"bar"}
// sniper job once foo 1 2 3 则 onceArgs = []string{"1", "2", "3"}
var onceArgs []string

var cmdOnce = &cobra.Command{
	Use:   "once job",
	Short: "Run job once",
	Long:  `Run job once.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		onceArgs = args[1:]
		if job, ok := jobs[name]; ok {
			job.job()
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
func addJob(name string, spec string, job func(ctx context.Context) error) {
	if _, ok := jobs[name]; ok {
		panic(name + "is used")
	}

	j := func() {
		ctx := context.Background()

		span, ctx := opentracing.StartSpanFromContext(ctx, "Cron")
		defer span.Finish()

		span.SetTag("name", name)
		ctx = ctxkit.WithTraceID(ctx, trace.GetTraceID(ctx))

		logger := log.Get(ctx)

		defer func() {
			if r := recover(); r != nil {
				logger.Error(r, string(debug.Stack()))
			}
		}()

		code := "0"
		t := time.Now()
		if err := job(ctx); err != nil {
			logger.Errorf("cron job error: %+v", err)
			code = "1"
		}
		d := time.Since(t)

		metrics.JobTotal.WithLabelValues(code).Inc()

		logger.WithField("cost", d.Seconds()).Infof("cron job %s[%s]", name, spec)
	}

	jobs[name] = jobInfo{spec: spec, job: j}

	if spec == "@manual" {
		return
	}

	if err := c.AddFunc(spec, j); err != nil {
		panic(err)
	}
}
