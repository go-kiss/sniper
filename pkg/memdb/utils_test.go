package memdb

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestSetOptions(t *testing.T) {
	opts := redis.Options{}
	dsn := "redis://foo:bar@localhost:6379/?DB=1&" +
		"&Network=tcp" +
		"&MaxRetries=1" +
		"&MinRetryBackoff=1s" +
		"&MaxRetryBackoff=1s" +
		"&DialTimeout=1s" +
		"&ReadTimeout=1s" +
		"&WriteTimeout=1s" +
		"&PoolFIFO=true" +
		"&PoolSize=1" +
		"&MinIdleConns=1" +
		"&MaxConnAge=1s" +
		"&PoolTimeout=1s" +
		"&IdleTimeout=1s" +
		"&IdleCheckFrequency=1s"

	setOptions(&opts, dsn)
	v := redis.Options{
		Network:            "tcp",
		Addr:               "localhost:6379",
		Username:           "foo",
		Password:           "bar",
		DB:                 1,
		MaxRetries:         1,
		MinRetryBackoff:    1 * time.Second,
		MaxRetryBackoff:    1 * time.Second,
		DialTimeout:        1 * time.Second,
		ReadTimeout:        1 * time.Second,
		WriteTimeout:       1 * time.Second,
		PoolFIFO:           true,
		PoolSize:           1,
		MinIdleConns:       1,
		MaxConnAge:         1 * time.Second,
		PoolTimeout:        1 * time.Second,
		IdleTimeout:        1 * time.Second,
		IdleCheckFrequency: 1 * time.Second,
	}

	if !reflect.DeepEqual(opts, v) {
		fmt.Println(opts)
	}
}
