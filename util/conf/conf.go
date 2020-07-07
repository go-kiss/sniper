// Package conf 提供最基础的配置加载功能
package conf

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var path string

var (
	// Hostname 主机名
	Hostname = "localhost"
	// AppID 获取 APP_ID
	AppID = "localapp"
	// IsDevEnv 开发环境标志
	IsDevEnv = false
	// IsUatEnv 集成环境标志
	IsUatEnv = false
	// IsProdEnv 生产环境标志
	IsProdEnv = false
	// Env 运行环境
	Env = "dev"
	// Zone 服务区域
	Zone = "sh001"
)

func init() {
	Hostname, _ = os.Hostname()
	if appID := os.Getenv("APP_ID"); appID != "" {
		AppID = appID
	} else {
		logger().Warn("env APP_ID is empty")
	}

	if env := os.Getenv("DEPLOY_ENV"); env != "" {
		Env = env
	} else {
		logger().Warn("env DEPLOY_ENV is empty")
	}

	if zone := os.Getenv("ZONE"); zone != "" {
		Zone = zone
	} else {
		logger().Warn("env ZONE is empty")
	}

	switch Env {
	case "prod", "pre":
		IsProdEnv = true
	case "uat":
		IsUatEnv = true
	default:
		IsDevEnv = true
	}

	path = os.Getenv("CONF_PATH")

	if path == "" {
		logger().Warn("env CONF_PATH is empty")
		var err error
		if path, err = os.Getwd(); err != nil {
			panic(err)
		}
		logger().WithField("path", path).Info("use default conf path")
	}


	viper.SetConfigName("sniper")

	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		logger().Error(err)
	}

	viper.AutomaticEnv()
}

// GetFloat64 获取浮点数配置
func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

// Get 获取字符串配置
func Get(key string) string {
	return viper.GetString(key)
}

// GetStrings 获取字符串列表
func GetStrings(key string) (s []string) {
	value := Get(key)
	if value == "" {
		return
	}

	for _, v := range strings.Split(value, ",") {
		s = append(s, v)
	}
	return
}

// GetInt32s 获取数字列表
// 1,2,3 => []int32{1,2,3}
func GetInt32s(key string) (s []int32, err error) {
	s64, err := GetInt64s(key)
	for _, v := range s64 {
		s = append(s, int32(v))
	}
	return
}

// GetInt64s 获取数字列表
func GetInt64s(key string) (s []int64, err error) {
	value := Get(key)
	if value == "" {
		return
	}

	var i int64
	for _, v := range strings.Split(value, ",") {
		i, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return
		}
		s = append(s, i)
	}
	return
}

// GetInt 获取整数配置
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetInt32 获取 int32 配置
func GetInt32(key string) int32 {
	return viper.GetInt32(key)
}

// GetInt64 获取 int64 配置
func GetInt64(key string) int64 {
	return viper.GetInt64(key)
}

// GetDuration 获取时间配置
func GetDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

// GetTime 查询时间配置
// 默认时间格式为 "2006-01-02 15:04:05"，conf.GetTime("FOO_BEGIN")
// 如果需要指定时间格式，则可以多传一个参数，conf.GetString("FOO_BEGIN", "2006")
//
// 配置不存在或时间格式错误返回**空时间对象**
// 使用本地时区
func GetTime(key string, args ...string) time.Time {
	fmt := "2006-01-02 15:04:05"
	if len(args) == 1 {
		fmt = args[0]
	}

	t, _ := time.ParseInLocation(fmt, viper.GetString(key), time.Local)
	return t
}

// GetBool 获取配置布尔配置
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// Set 设置配置，仅用于测试
func Set(key string, value string) {
	viper.Set(key, value)
}

// OnConfigChange 注册配置文件变更回调
// 需要在 WatchConfig 之前调用
func OnConfigChange(run func()) {
	viper.OnConfigChange(func(in fsnotify.Event) { run() })
}

// WatchConfig 启动配置变更监听，业务代码不要调用。
func WatchConfig() {
	viper.WatchConfig()
}

var levels = map[string]logrus.Level{
	"panic": logrus.PanicLevel,
	"fatal": logrus.FatalLevel,
	"error": logrus.ErrorLevel,
	"warn":  logrus.WarnLevel,
	"info":  logrus.InfoLevel,
	"debug": logrus.DebugLevel,
}

func logger() *logrus.Entry {
	if level, ok := levels[os.Getenv("LOG_LEVEL")]; ok {
		logrus.SetLevel(level)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}

	return logrus.WithFields(logrus.Fields{
		"app_id":      AppID,
		"instance_id": Hostname,
		"env":         Env,
	})
}
