package sotatgbot

import (
	"context"
	"fmt"
	"github.com/CookieNyanCloud/tgFeedBackBot/configs"
	"github.com/CookieNyanCloud/tgFeedBackBot/repository"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	redErr  = "error in redis:%v\n"
	welcome = "Привет, я связующая бездна"
	Next    = "Вперед"
)

type Actions struct {
	Cache *repository.Repo
	Bot   *tgbotapi.BotAPI
	Ctx   context.Context
	Cfg   *configs.Conf
}

func NewActions(
	cache *repository.Repo,
	bot *tgbotapi.BotAPI,
	ctx context.Context,
	cfg *configs.Conf) *Actions {
	return &Actions{
		Cache: cache,
		Bot:   bot,
		Ctx:   ctx,
		Cfg:   cfg,
	}
}

type ActionsInterface interface {
	ReplyToMsg(chatId int, txt string)
	SendMsg(chatId int64, msgId int)
	BanUser(msgId int)
	CheckBanUser(chatId int64) bool
}

func (a *Actions) StartMsg(chatId int64) {
	msg := tgbotapi.NewMessage(chatId, welcome)
	_, _ = a.Bot.Send(msg)
}

func (a *Actions) ReplyToMsg(chatId int, txt string) {
	id, err := a.Cache.GetUser(a.Ctx, chatId)
	if err != nil && err != redis.Nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
		return
	} else if err == redis.Nil {
		fmt.Println("no user")
		return
	}
	msg := tgbotapi.NewMessage(id, txt)
	_, _ = a.Bot.Send(msg)
}

func (a *Actions) SendMsg(chatId int64, msgId int) {

	msg := tgbotapi.ForwardConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: a.Cfg.Chat,
		},
		FromChatID: chatId,
		MessageID:  msgId,
	}
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{
		HideKeyboard: false,
		Selective:    false,
	}
	forwarded, _ := a.Bot.Send(msg)
	err := a.Cache.SetUser(a.Ctx, chatId, forwarded.MessageID)
	if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
	}
}

func (a *Actions) BanUser(msgId int) {
	id, err := a.Cache.GetUser(a.Ctx, msgId)
	if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
		return
	}
	err = a.Cache.SetBan(a.Ctx, id)
	if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
		return
	}
}

func (a *Actions) CheckBanUser(chatId int64) bool {
	state, err := a.Cache.GetBan(a.Ctx, chatId)
	if err == redis.Nil {
		return false
	} else if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
		return false
	}
	return state
}
