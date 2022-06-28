package redisDB

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Repo struct {
	db *redis.Client
}

func NewRepo(db *redis.Client) *Repo {
	return &Repo{db: db}
}

type UsersInterface interface {
	SetUser(ctx context.Context, userId int64, msgId int) error
	GetUser(ctx context.Context, msgId int) (int64, error)
	SetBan(ctx context.Context, userId int64) error
	GetBan(ctx context.Context, userId int64) (bool, error)
}

func (r *Repo) SetUser(ctx context.Context, userId int64, msgId int) error {
	return r.db.Set(ctx, strconv.Itoa(msgId), userId, time.Hour*168).Err()
}

func (r *Repo) GetUser(ctx context.Context, msgId int) (int64, error) {
	idStr, err := r.db.Get(ctx, strconv.Itoa(msgId)).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(idStr, 10, 64)
}

func (r *Repo) SetBan(ctx context.Context, userId int64) error {
	idStr := strconv.FormatInt(userId, 10)
	return r.db.Set(ctx, idStr, true,0).Err()
}

func (r *Repo) GetBan(ctx context.Context, userId int64) (bool, error) {
	idStr := strconv.FormatInt(userId, 10)
	return r.db.Get(ctx, idStr).Bool()
}
