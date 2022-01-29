package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/CookieNyanCloud/tgFeedBackBot/configs"
	"github.com/CookieNyanCloud/tgFeedBackBot/repository/database/redisDB"
	"github.com/CookieNyanCloud/tgFeedBackBot/sotatgbot"
	"github.com/go-redis/redis/v8"

	"syscall"
)

const (
	dbErr      = "error init redis:%v\n"
	confErr    = "error init conf: %v\n"
	closeDbErr = "error closing db: %v\n"
)

func main() {

	var ctx = context.Background()
	conf, err := configs.InitConf()
	if err != nil {
		log.Fatalf(confErr, err)
	}

	redisClient, err := redisDB.NewDatabase(conf.Redis, ctx)
	if err != nil {
		log.Fatalf(dbErr, err)
	}
	cache := redisDB.NewRepo(redisClient.Client)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func(ctx context.Context, db *redis.Client) {
		<-quit
		fmt.Println("shutdown")
		const timeout = 5 * time.Second
		ctx, shutdown := context.WithTimeout(context.Background(), timeout)
		defer shutdown()
		if err := db.Close(); err != nil {
			log.Fatalf(closeDbErr, err)
		}
		os.Exit(1)
	}(ctx, redisClient.Client)

	bot, updates := sotatgbot.StartSotaBot(conf.Token)
	act := sotatgbot.NewActions(ctx, cache, bot, conf)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Command() == "start" {
			act.StartMsg(update.Message.Chat.ID)
			continue
		}

		if act.CheckBanUser(update.Message.Chat.ID) {
			continue
		}

		if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage != nil {
			if update.Message.Command() == "ban" {
				act.BanUser(update.Message.ReplyToMessage.MessageID)
				continue
			} else {
				act.ReplyToMsg(update.Message.ReplyToMessage.MessageID, update.Message.Text)
				continue
			}
		} else if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage == nil {
			continue
		}
		act.SendMsg(update.Message.Chat.ID, update.Message.MessageID)
	}
}
