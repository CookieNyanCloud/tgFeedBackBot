package main

import (
	"context"
	"fmt"
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
	none    = "Не знаю такой команды"
	needTxt = "Добавьте текст к запросу, пожалуйста"
	dbErr   = "что-то с базой:%s"

	welcome      = "Привет, я связующая бездна"
	next         = "Вперед"
	back1        = "Назад в меню"
	back2        = "Назад"
	help         = "Помочь Соте"
	msgNearStart = "Меню стартует здесь!"
	tell         = "Рассказать о чем-то очень важном"
	telltext     = "Таки да?"
	helpText     = `Не хлебом единым! Или хлебом?


Помочь Соте можно здесь: 
https://donationalerts.ru/r/sota_vision 
https://www.patreon.com/sotavision
https://boosty.to/sota
Карта Сбербанка: 4817760237727932 
Яндекс.Кошелек: 41001502944105 
PayPal: sotadonation@gmail.com 
WebMoney: Z207958641009
BTC: bc1qfumf299wacxh4r8djnevxsc9xuj4rk8wf6sjwe
ETC: 0x18ADb185fD627737Cb2458f0D6037F596D167f38`
)

func main() {
	var ctx = context.Background()
	conf, err := configs.InitConf()
	if err != nil {
		log.Fatalf("error getting users: %v", err)
	}

	redisClient, err := redisDB.NewDatabase(conf.Redis, ctx)
	if err != nil {
		log.Fatalf("error init db: %v", err)
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
			log.Fatalf("error closing db: %v", err)
		}
		os.Exit(1)

	}(ctx, redisClient.Client)

	bot, updates := sotatgbot.StartSotaBot(conf.Token)
	for update := range updates {

		keyboard := tgbotapi.ReplyKeyboardMarkup{}

		if update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcome)
			keyboard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(next)))
			msg.ReplyMarkup = keyboard
			err := cache.SetState(ctx, update.Message.Chat.ID, false)
			if err != nil {
				msgtext := fmt.Sprintf(dbErr, err)
				msg := tgbotapi.NewMessage(conf.Chat, msgtext)
				_, _ = bot.Send(msg)
				continue
			}
			_, _ = bot.Send(msg)
			continue
		}

		if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage != nil {
			id, err := cache.GetUser(ctx, update.Message.ReplyToMessage.MessageID)
			if err != nil {
				msgtext := fmt.Sprintf(dbErr, err)
				msg := tgbotapi.NewMessage(conf.Chat, msgtext)
				_, _ = bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(id, update.Message.Text)
			_, _ = bot.Send(msg)
			continue
		} else if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage == nil {
			continue
		}

		switch update.Message.Text {

		case next, back1, back2:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgNearStart)
			keyBoard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(tell),
					tgbotapi.NewKeyboardButton(help),
				))
			msg.ReplyMarkup = keyBoard
			err := cache.SetState(ctx, update.Message.Chat.ID, false)
			if err != nil {
				msgtext := fmt.Sprintf(dbErr, err)
				msg := tgbotapi.NewMessage(conf.Chat, msgtext)
				_, _ = bot.Send(msg)
				continue
			}
			_, _ = bot.Send(msg)

		case help:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
			keyBoard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(back1),
				))
			msg.ReplyMarkup = keyBoard
			err := cache.SetState(ctx, update.Message.Chat.ID, false)
			if err != nil {
				msgtext := fmt.Sprintf(dbErr, err)
				msg := tgbotapi.NewMessage(conf.Chat, msgtext)
				_, _ = bot.Send(msg)
				continue
			}
			_, _ = bot.Send(msg)

		case tell:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, telltext)
			keyBoard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(back2),
				))
			err := cache.SetState(ctx, update.Message.Chat.ID, true)
			if err != nil {
				msgtext := fmt.Sprintf(dbErr, err)
				msg := tgbotapi.NewMessage(conf.Chat, msgtext)
				_, _ = bot.Send(msg)
				continue
			}
			msg.ReplyMarkup = keyBoard
			_, _ = bot.Send(msg)

		default:
			state, err := cache.GetState(ctx, update.Message.Chat.ID)
			if err != nil {
				msgtext := fmt.Sprintf(dbErr, err)
				msg := tgbotapi.NewMessage(conf.Chat, msgtext)
				_, _ = bot.Send(msg)
				continue
			}
			if state {
				err := cache.SetState(ctx, update.Message.Chat.ID, true)
				if err != nil {
					msgtext := fmt.Sprintf(dbErr, err)
					msg := tgbotapi.NewMessage(conf.Chat, msgtext)
					_, _ = bot.Send(msg)
					continue
				}
				msg := tgbotapi.ForwardConfig{
					BaseChat: tgbotapi.BaseChat{
						ChatID: conf.Chat,
					},
					FromChatID: update.Message.Chat.ID,
					MessageID:  update.Message.MessageID,
				}
				msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
					HideKeyboard: false,
					Selective:    false,
				}
				forwarded, _ := bot.Send(msg)
				err = cache.SetUser(ctx, update.Message.Chat.ID, forwarded.MessageID)
				if err != nil {
					msgtext := fmt.Sprintf(dbErr, err)
					msg := tgbotapi.NewMessage(conf.Chat, msgtext)
					_, _ = bot.Send(msg)
				}
			} else {
				err := cache.SetState(ctx, update.Message.Chat.ID, false)
				if err != nil {
					msgtext := fmt.Sprintf(dbErr, err)
					msg := tgbotapi.NewMessage(conf.Chat, msgtext)
					_, _ = bot.Send(msg)
					continue
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, none)
				_, _ = bot.Send(msg)
			}

		}
	}
}
