package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

type Client struct {
	*redis.Client
}

type Config struct {
	Host         string // redis name, for trace
	Password     string
	Db           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func getConfig() *Config {
	return &Config{
		Host:         viper.GetString("redis.host"),
		Password:     viper.GetString("redis.password"),
		Db:           viper.GetInt("redis.db"),
		DialTimeout:  viper.GetDuration("redis.dial_timeout"),
		ReadTimeout:  viper.GetDuration("redis.read_timeout"),
		WriteTimeout: viper.GetDuration("redis.write_timeout"),
	}
}

func NewRedis() *Client {
	conf := getConfig()
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Host,
		Password: conf.Password,
		DB:       conf.Db,
		//DialTimeout:  conf.DialTimeout * time.Second,
		//ReadTimeout:  conf.ReadTimeout,
		//WriteTimeout: conf.WriteTimeout,
		//PoolSize:    10,
		//PoolTimeout: 30 * time.Second,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}

	return &Client{Client: rdb}
}

func (c *Client) Close() error {
	return c.Client.Close()
}
