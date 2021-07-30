# cmd/cron

注册定时任务请参考 [demo.go](./demo.go)。

查看所有定时任务
```bash
go main.go job list
```

执行一次某个任务
```bash
go main.go job once foo
```

调度所有定时任务
```bash
go main.go job
```
