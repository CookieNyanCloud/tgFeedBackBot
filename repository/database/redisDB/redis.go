package redisDB

import (
	"context"
	"errors"
	"github.com/CookieNyanCloud/tgFeedBackBot/configs"
	"github.com/go-redis/redis/v8"
)

var (
	ErrNil = errors.New("no matching record found in redisDB database")
)

type Database struct {
	Client *redis.Client
}

func NewDatabase(cfg configs.RedisConf, ctx context.Context) (*Database, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Database{
		Client: client,
	}, nil
}
