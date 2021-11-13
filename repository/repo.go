package repository

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
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
	err := r.db.Set(ctx, string(msgId), userId, time.Hour*24).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetUser(ctx context.Context, msgId int) (int64, error) {
	idStr, err := r.db.Get(ctx, string(msgId)).Result()
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		panic(err)
	}
	return id, err
}

func (r *Repo) SetBan(ctx context.Context, userId int64) error {
	idStr := strconv.FormatInt(userId, 10)
	r.db.Set(ctx, "ban_"+idStr, true, time.Hour*100)
	return nil
}

func (r *Repo) GetBan(ctx context.Context, userId int64) (bool, error) {
	idStr := strconv.FormatInt(userId, 10)
	stateStr, err := r.db.Get(ctx, "ban_"+idStr).Result()

	if err != nil {
		return false, err
	}
	state, err := strconv.ParseBool(stateStr)
	if err != nil {
		return false, err
	}
	return state, err
}
