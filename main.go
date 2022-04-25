package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"syscall"

	"github.com/CookieNyanCloud/tgFeedBackBot/configs"
	"github.com/CookieNyanCloud/tgFeedBackBot/repository/database/redisDB"
	"github.com/CookieNyanCloud/tgFeedBackBot/sotatgbot"
	"github.com/go-redis/redis/v8"
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
			} else if update.Message.Command() == "idsend" {
				args := make([]string, 2)
				args = strings.Split(update.Message.CommandArguments(), "%")
				id, err := strconv.Atoi(args[0])
				if err != nil {
					fmt.Println(err)
					continue
				}
				act.ReplyToMsgTxtById(int64(id), args[1])
				continue
			} else {
				if update.Message.Text != "" {
					act.ReplyToMsgTxt(update.Message.ReplyToMessage.MessageID, update.Message.Text)
					continue
				}
				if update.Message.Document != nil {
					act.ReplyToMsgFile(update.Message.ReplyToMessage.MessageID, update.Message.Document.FileID)
					continue
				}
				if update.Message.Photo != nil {
					act.ReplyToMsgPhotoVideo(update.Message.ReplyToMessage.MessageID, update.Message.Photo[len(update.Message.Photo)-1].FileID, "photo", update.Message.Caption)
					continue
				}
				if update.Message.Video != nil {
					act.ReplyToMsgPhotoVideo(update.Message.ReplyToMessage.MessageID, update.Message.Video.FileID, "video", update.Message.Caption)
					continue
				}
				if update.Message.Sticker != nil {
					act.ReplyToMsgMedia(update.Message.ReplyToMessage.MessageID, update.Message.Sticker.FileID, "sticker")
					continue
				}
				if update.Message.Voice != nil {
					act.ReplyToMsgMedia(update.Message.ReplyToMessage.MessageID, update.Message.Voice.FileID, "voice")
					continue
				}
				continue
			}
		} else if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage == nil {
			continue
		}
		act.SendMsg(update.Message.Chat.ID, update.Message.MessageID)
	}
}
