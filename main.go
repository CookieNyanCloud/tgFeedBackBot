package main

import (
	"fmt"
	"github.com/CookieNyanCloud/tgFeedBackBot/configs"
	"github.com/CookieNyanCloud/tgFeedBackBot/repository"
	"github.com/CookieNyanCloud/tgFeedBackBot/repository/database/postgres"
	"github.com/CookieNyanCloud/tgFeedBackBot/sotatgbot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"

	"syscall"
)

const (
	welcome      = "Привет, я связующая бездна"
	next         = "Вперед"
	back1        = "Назад в меню"
	back2        = "Назад"
	help         = "Помочь Соте"
	msgNearStart = "Меню стартует здесь!"
	tell         = "Рассказать о чем-то очень важном"
	telltext     = "Таки да?"
	none         = "Не знаю такой команды"
	work         = "Работаем"
	needTxt      = "Добавьте текст к запросу, пожалуйста"
	needIndex    = "индекс для ответа:%d"
	dbErr        = "что-то с базой:%s"
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

	conf := configs.InitConf()
	users, err := configs.InitUsers()
	if err != nil {
		log.Fatalf("error getting users: %v", err)
	}

	postgresClient, err := postgres.NewClient(conf.Postgres)
	if err != nil {
		log.Fatalf("error init db: %v", err)
	}
	repos := repository.NewRepo(postgresClient)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func(users map[int64]bool, db *sqlx.DB) {
		<-quit
		err := configs.SaveUsers(users)
		if err != nil {
			log.Fatalf("error getting users: %v", err)
		}
		if err := db.Close(); err != nil {
			log.Fatalf("error closing db: %v", err)
		}
		os.Exit(1)
	}(users, postgresClient)

	bot, updates := sotatgbot.StartSotaBot(conf.Token)
	for update := range updates {

		keyboard := tgbotapi.ReplyKeyboardMarkup{}
		if update.Message == nil {
			continue
		}
		if update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcome)
			keyboard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(next)))
			msg.ReplyMarkup = keyboard
			users[update.Message.Chat.ID] = false
			_, _ = bot.Send(msg)
			continue
		}


		if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage != nil {

			var id int64
			if update.Message.ReplyToMessage.ForwardFrom != nil {
				id = int64(update.Message.ReplyToMessage.ForwardFrom.ID)

			} else {
				var txt string
				if update.Message.ReplyToMessage.Text != "" {
					txt = update.Message.ReplyToMessage.Text

				} else if update.Message.ReplyToMessage.Caption != "" {
					txt = update.Message.ReplyToMessage.Caption
				} else {
					txt = ""
				}

				id, err = repos.GetId(txt, update.Message.ReplyToMessage.ForwardDate)
				if err != nil {
					msgtext := fmt.Sprintf(dbErr, err)
					msg := tgbotapi.NewMessage(conf.Chat, msgtext)
					_, _ = bot.Send(msg)
				}
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
			users[update.Message.Chat.ID] = false
			_, _ = bot.Send(msg)

		case help:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
			keyBoard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(back1),
				))
			msg.ReplyMarkup = keyBoard
			users[update.Message.Chat.ID] = false
			_, _ = bot.Send(msg)

		case tell:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, telltext)
			keyBoard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(back2),
				))
			users[update.Message.Chat.ID] = true
			msg.ReplyMarkup = keyBoard
			_, _ = bot.Send(msg)

		default:
			if users[update.Message.Chat.ID] {
				users[update.Message.Chat.ID] = true
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
				var txt string
				if update.Message.Text != "" {
					txt = update.Message.Text
				} else if update.Message.Caption != ""{
					txt = update.Message.Caption
				} else {
					//txt = fmt.Sprintf(needIndex,rand.Int())
					txt = ""
				}
				_, _ = bot.Send(msg)
				err = repos.MakeSms(update.Message.Chat.ID, txt, update.Message.Date)
				if err != nil {
					msgtext := fmt.Sprintf(dbErr, err)
					msg := tgbotapi.NewMessage(conf.Chat, msgtext)
					_, _ = bot.Send(msg)
				}
			} else if update.Message.Caption == "" || update.Message.Text == "" {
				users[update.Message.Chat.ID] = false
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, none)
				_, _ = bot.Send(msg)
			} else {
				users[update.Message.Chat.ID] = false
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, none)
				_, _ = bot.Send(msg)
			}

		}
	}
}
