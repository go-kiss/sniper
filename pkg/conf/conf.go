// Package conf 提供最基础的配置加载功能
package conf

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Get 查询配置/环境变量
var Get func(string) string

var (
	// Host 主机名
	Host = "localhost"
	// App 服务标识
	App = "localapp"
	// Env 运行环境
	Env = "dev"
	// Zone 服务区域
	Zone = "sh001"

	files = map[string]*Conf{}

	defaultFile = "sniper"
)

func init() {
	Host, _ = os.Hostname()
	if appID := os.Getenv("APP_ID"); appID != "" {
		App = appID
	}

	if env := os.Getenv("ENV"); env != "" {
		Env = env
	}

	if zone := os.Getenv("ZONE"); zone != "" {
		Zone = zone
	}

	if name := os.Getenv("CONF_NAME"); name != "" {
		defaultFile = name
	}

	path := os.Getenv("CONF_PATH")
	if path == "" {
		var err error
		if path, err = os.Getwd(); err != nil {
			panic(err)
		}
	}

	fs, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, f := range fs {
		if !strings.HasSuffix(f.Name(), ".toml") {
			continue
		}

		v := viper.New()
		v.SetConfigFile(filepath.Join(path, f.Name()))
		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}
		v.AutomaticEnv()

		name := strings.TrimSuffix(f.Name(), ".toml")
		files[name] = &Conf{v}
	}

	Get = GetString
}

type Conf struct {
	*viper.Viper
}

// File 根据文件名获取对应配置对象
// 目前仅支持 toml 文件，不用传扩展名
// 如果要读取 foo.toml 配置，可以 File("foo").Get("bar")
func File(name string) *Conf {
	return files[name]
}

// OnConfigChange 注册配置文件变更回调
// 需要在 WatchConfig 之前调用
func OnConfigChange(run func()) {
	for _, v := range files {
		v.OnConfigChange(func(in fsnotify.Event) { run() })
	}
}

// WatchConfig 启动配置变更监听，业务代码不要调用。
func WatchConfig() {
	for _, v := range files {
		v.WatchConfig()
	}
}

// Set 设置配置，仅用于测试
func Set(key string, value interface{}) { File(defaultFile).Set(key, value) }

func GetBool(key string) bool              { return File(defaultFile).GetBool(key) }
func GetDuration(key string) time.Duration { return File(defaultFile).GetDuration(key) }
func GetFloat64(key string) float64        { return File(defaultFile).GetFloat64(key) }
func GetInt(key string) int                { return File(defaultFile).GetInt(key) }
func GetInt32(key string) int32            { return File(defaultFile).GetInt32(key) }
func GetInt64(key string) int64            { return File(defaultFile).GetInt64(key) }
func GetIntSlice(key string) []int         { return File(defaultFile).GetIntSlice(key) }
func GetSizeInBytes(key string) uint       { return File(defaultFile).GetSizeInBytes(key) }
func GetString(key string) string          { return File(defaultFile).GetString(key) }
func GetStringSlice(key string) []string   { return File(defaultFile).GetStringSlice(key) }
func GetTime(key string) time.Time         { return File(defaultFile).GetTime(key) }
func GetUint(key string) uint              { return File(defaultFile).GetUint(key) }
func GetUint32(key string) uint32          { return File(defaultFile).GetUint32(key) }
func GetUint64(key string) uint64          { return File(defaultFile).GetUint64(key) }

func GetStringMap(key string) map[string]interface{} { return File(defaultFile).GetStringMap(key) }
func GetStringMapString(key string) map[string]string {
	return File(defaultFile).GetStringMapString(key)
}
func GetStringMapStringSlice(key string) map[string][]string {
	return File(defaultFile).GetStringMapStringSlice(key)
}
