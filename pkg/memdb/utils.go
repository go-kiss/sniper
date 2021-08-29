package memdb

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func setOptions(opts *redis.Options, dsn string) {
	url, err := url.Parse(dsn)
	if err != nil {
		panic(err)
	}

	args := url.Query()

	rv := reflect.ValueOf(opts).Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		f := rv.Field(i)
		if !f.CanInterface() {
			continue
		}
		name := rt.Field(i).Name
		arg := args.Get(name)
		if arg == "" {
			continue
		}
		switch f.Interface().(type) {
		case time.Duration:
			v, err := time.ParseDuration(arg)
			if err != nil {
				panic(fmt.Sprintf("%s=%s, err:%v", name, arg, err))
			}
			f.Set(reflect.ValueOf(v))
		case int:
			v, err := strconv.Atoi(arg)
			if err != nil {
				panic(fmt.Sprintf("%s=%s, err:%v", name, arg, err))
			}
			f.SetInt(int64(v))
		case bool:
			v, err := strconv.ParseBool(arg)
			if err != nil {
				panic(fmt.Sprintf("%s=%s, err:%v", name, arg, err))
			}
			f.SetBool(v)
		case string:
			f.SetString(arg)
		}
	}

	opts.Addr = url.Host
	opts.Username = url.User.Username()
	if p, ok := url.User.Password(); ok {
		opts.Password = p
	}
}

func name(ctx context.Context) string {
	v, _ := ctx.Value(nameKey{}).(string)
	if v == "" {
		v = "unknown"
	}
	return v
}
