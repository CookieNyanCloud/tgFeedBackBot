package sotatgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func StartSotaBot(token string) (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 600
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		println(err)
	}
	return bot, updates
}
