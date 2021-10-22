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

type Users struct {
	MsgID int64
	State bool
}


type UsersInterface interface {
	SetUser(ctx context.Context, userId int64, msgId int ) error
	GetUser(ctx context.Context, msgId int) (int64, error)
	SetState(ctx context.Context, userId int64, state bool) error
	GetState(ctx context.Context, userId int) (bool, error)
	SetBan(ctx context.Context, userId int64) error
	GetBan(ctx context.Context, userId int) (bool, error)
}

func (r *Repo) SetUser(ctx context.Context, userId int64, msgId int) error {
	return r.db.Set(ctx,string(msgId),userId,time.Hour*24).Err()
}

func (r *Repo) GetUser(ctx context.Context, msgId int) (int64, error) {
	idStr, err:=r.db.Get(ctx,string(msgId)).Result()
	if err != nil {
		return 0,err
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		panic(err)
	}
	return id,err
}

func (r *Repo) SetState(ctx context.Context, userId int64, state bool) error {
	r.db.Set(ctx,string(userId),state,time.Hour*24)
	return nil
}

func (r *Repo)GetState(ctx context.Context, userId int64) (bool, error) {
	stateStr, err:=r.db.Get(ctx,string(userId)).Result()
	if err != nil {
		return false,err
	}
	state, err:= strconv.ParseBool(stateStr)
	if err != nil {
		return false,err
	}
	return state,err
}

func (r *Repo) SetBan(ctx context.Context, userId int64) error {
	r.db.Set(ctx,"ban_"+string(userId),true,time.Hour*100)
	return nil
}

func (r *Repo)GetBan(ctx context.Context, userId int64) (bool, error) {
	stateStr, err:=r.db.Get(ctx,"ban_"+string(userId)).Result()
	if err != nil {
		return false,err
	}
	state, err:= strconv.ParseBool(stateStr)
	if err != nil {
		return false,err
	}
	return state,err
}
