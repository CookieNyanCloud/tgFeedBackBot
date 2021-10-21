package main

import (
	"context"
	"github.com/CookieNyanCloud/tgFeedBackBot/configs"
	"github.com/CookieNyanCloud/tgFeedBackBot/repository"
	"github.com/CookieNyanCloud/tgFeedBackBot/repository/database/redisDB"
	"github.com/CookieNyanCloud/tgFeedBackBot/sotatgbot"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"os/signal"
	"time"

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
	cache := repository.NewRepo(redisClient.Client)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func(ctx context.Context, db *redis.Client) {
		<-quit
		const timeout = 5 * time.Second
		ctx, shutdown := context.WithTimeout(context.Background(), timeout)
		defer shutdown()
		if err := db.Close(); err != nil {
			log.Fatalf(closeDbErr, err)
		}
		os.Exit(1)
	}(ctx, redisClient.Client)

	bot, updates := sotatgbot.StartSotaBot(conf.Token)
	keyboard := tgbotapi.ReplyKeyboardMarkup{}
	act := sotatgbot.NewActions(cache, bot, ctx, keyboard, conf)
	for update := range updates {
		if update.Message.Command() == "start" {
			act.StartMsg(update.Message.Chat.ID)
			continue
		}
		if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage != nil {
			act.ReplyToMsg(update.Message.ReplyToMessage.MessageID, update.Message.Text)
			continue
		} else if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage == nil {
			continue
		}

		switch update.Message.Text {

		case sotatgbot.Next, sotatgbot.Back1, sotatgbot.Back2:
			act.NextBack(update.Message.Chat.ID)

		case sotatgbot.Help:
			act.HelpMsg(update.Message.Chat.ID)

		case sotatgbot.Tell:
			act.TellMsg(update.Message.Chat.ID)

		default:
			act.SendMsg(update.Message.Chat.ID, update.Message.MessageID)

		}
	}
}
