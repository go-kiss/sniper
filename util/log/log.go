// Package log 基础日志组件
package log

import (
	"context"
	"os"
	"strconv"

	"sniper/util/conf"
	"sniper/util/ctxkit"

	"github.com/k0kubun/pp"
	isatty "github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

func init() {
	setLevel()
	initPP()
}

func initPP() {
	out := os.Stdout
	pp.SetDefaultOutput(out)

	if !isatty.IsTerminal(out.Fd()) {
		pp.ColoringEnabled = false
	}
}

// Logger logger
type Logger = *logrus.Entry

// Fields fields
type Fields = logrus.Fields

var levels = map[string]logrus.Level{
	"panic": logrus.PanicLevel,
	"fatal": logrus.FatalLevel,
	"error": logrus.ErrorLevel,
	"warn":  logrus.WarnLevel,
	"info":  logrus.InfoLevel,
	"debug": logrus.DebugLevel,
}

func setLevel() {
	levelConf := conf.Get("LOG_LEVEL_" + conf.Hostname)

	if levelConf == "" {
		levelConf = conf.Get("LOG_LEVEL")
	}

	if level, ok := levels[levelConf]; ok {
		logrus.SetLevel(level)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

// Get 获取日志实例
func Get(ctx context.Context) Logger {
	return logrus.WithFields(logrus.Fields{
		"env":         conf.Env,
		"app_id":      conf.AppID,
		"instance_id": conf.Hostname,
		"ip":          ctxkit.GetUserIP(ctx),
		"trace_id":    ctxkit.GetTraceID(ctx),
}

// Reset 使用最新配置重置日志级别
func Reset() {
	setLevel()
}

// PP 类似 PHP 的 var_dump
func PP(args ...interface{}) {
	pp.Println(args...)
}
