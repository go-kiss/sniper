module sniper

go 1.16

replace sniper/pkg => ./pkg/

require (
	github.com/dave/dst v0.25.5
	github.com/opentracing/opentracing-go v1.2.0
	github.com/prometheus/client_golang v1.11.0
	github.com/robfig/cron v1.2.0
	github.com/spf13/cobra v1.2.1
	go.uber.org/automaxprocs v1.4.0
	google.golang.org/protobuf v1.27.1
	sniper/pkg v0.0.0-00010101000000-000000000000
)

require github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
