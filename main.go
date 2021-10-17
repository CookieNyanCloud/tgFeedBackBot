package main

import (
	"github.com/CookieNyanCloud/tgFeedBackBot/configs"
	"github.com/CookieNyanCloud/tgFeedBackBot/sotatgbot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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
	helpText     = `Не хлебом единым! Или хлебом?

Помочь Соте рублём можно здесь: 
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	conf := configs.InitConf()
	bot, updates := sotatgbot.StartSotaBot(conf.Token)
	users, err := configs.InitUsers()
	go func(users map[int64]bool) {
		<-quit
		err := configs.SaveUsers(users)
		if err != nil {
			log.Fatalf("error getting users: %v", err)
		}
		os.Exit(1)

	}(users)
	if err != nil {
		log.Fatalf("error getting users: %v", err)
	}

	for update := range updates {

		keyboard := tgbotapi.ReplyKeyboardMarkup{}

		if update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcome)
			keyboard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(next)))
			msg.ReplyMarkup = keyboard
			users[update.Message.Chat.ID] = false
			_, _ = bot.Send(msg)
			continue
		}

		if update.Message.Chat.ID == conf.Chat && update.Message.ReplyToMessage != nil {
			msg := tgbotapi.NewMessage(int64(update.Message.From.ID), update.Message.Text)
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
				_, _ = bot.Send(msg)
				//msg2:=tgbotapi.NewMessage(update.Message.Chat.ID,work)
				//_, _ = bot.Send(msg2)
			} else {
				users[update.Message.Chat.ID] = false
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, none)
				_, _ = bot.Send(msg)
			}

		}
	}
}
