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
	redErr       = "error in redis:%v\n"
	banErr       = "праблема с баном: %v\n"
	welcome      = "Привет, я связующая бездна"
	none         = "Не знаю такой команды"
	Next         = "Вперед"
	msgNearStart = "Меню стартует здесь!"
	Help         = "Помочь Соте"
	Tell         = "Рассказать о чем-то очень важном"
	Back1        = "Назад в меню"
	Back2        = "Назад"

	telltext = "Таки да?"
	helpText = `Не хлебом единым! Или хлебом?
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

type Actions struct {
	Cache    *repository.Repo
	Bot      *tgbotapi.BotAPI
	Ctx      context.Context
	Keyboard tgbotapi.ReplyKeyboardMarkup
	Cfg      *configs.Conf
}

func NewActions(
	cache *repository.Repo,
	bot *tgbotapi.BotAPI,
	ctx context.Context,
	keyboard tgbotapi.ReplyKeyboardMarkup,
	cfg *configs.Conf) *Actions {
	return &Actions{
		Cache:    cache,
		Bot:      bot,
		Ctx:      ctx,
		Keyboard: keyboard,
		Cfg:      cfg,
	}
}

type ActionsInterface interface {
	StartMsg(chatId int64)
	ReplyToMsg(chatId int, txt string)
	NextBack(chatId int64)
	HelpMsg(chatId int64)
	TellMsg(chatId int64)
	SendMsg(chatId int64, msgId int)
	BanUser(msgId int)
	CheckBanUser(chatId int64)
}

func (a *Actions) StartMsg(chatId int64) {
	msg := tgbotapi.NewMessage(chatId, welcome)
	a.Keyboard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Next)))
	msg.ReplyMarkup = a.Keyboard
	err := a.Cache.SetState(a.Ctx, chatId, false)
	if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
		return
	}
	_, _ = a.Bot.Send(msg)
}

func (a *Actions) ReplyToMsg(chatId int, txt string) {
	id, err := a.Cache.GetUser(a.Ctx, chatId)
	if err == redis.Nil {
		//msgtext := fmt.Sprintf(redErr, err)
		//msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		//_, _ = a.Bot.Send(msg)
		return
	} else if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(id, txt)
	_, _ = a.Bot.Send(msg)
}

func (a *Actions) NextBack(chatId int64) {
	msg := tgbotapi.NewMessage(chatId, msgNearStart)
	a.Keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Tell),
			tgbotapi.NewKeyboardButton(Help),
		))
	msg.ReplyMarkup = a.Keyboard
	err := a.Cache.SetState(a.Ctx, chatId, false)
	if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
	}
	_, _ = a.Bot.Send(msg)
}

func (a *Actions) HelpMsg(chatid int64) {
	msg := tgbotapi.NewMessage(chatid, helpText)
	a.Keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Back1),
		))
	msg.ReplyMarkup = a.Keyboard
	err := a.Cache.SetState(a.Ctx, chatid, false)
	if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
	}
	_, _ = a.Bot.Send(msg)
}

func (a *Actions) TellMsg(chatId int64) {
	msg := tgbotapi.NewMessage(chatId, telltext)
	a.Keyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Back2),
		))
	err := a.Cache.SetState(a.Ctx, chatId, true)
	if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
	}
	msg.ReplyMarkup = a.Keyboard
	_, _ = a.Bot.Send(msg)
}

func (a *Actions) SendMsg(chatId int64, msgId int) {
	state, err := a.Cache.GetState(a.Ctx, chatId)
	if err != nil {
		msgtext := fmt.Sprintf(redErr, err)
		msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
		_, _ = a.Bot.Send(msg)
	}
	if state {
		err := a.Cache.SetState(a.Ctx, chatId, true)
		if err != nil {
			msgtext := fmt.Sprintf(redErr, err)
			msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
			_, _ = a.Bot.Send(msg)
		}
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
		err = a.Cache.SetUser(a.Ctx, chatId, forwarded.MessageID)
		if err != nil {
			msgtext := fmt.Sprintf(redErr, err)
			msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
			_, _ = a.Bot.Send(msg)
		}
	} else {
		err := a.Cache.SetState(a.Ctx, chatId, false)
		if err != nil {
			msgtext := fmt.Sprintf(redErr, err)
			msg := tgbotapi.NewMessage(a.Cfg.Chat, msgtext)
			_, _ = a.Bot.Send(msg)
		}
		msg := tgbotapi.NewMessage(chatId, none)
		_, _ = a.Bot.Send(msg)
	}
}

func (a *Actions) BanUser(msgId int) {
	id, err := a.Cache.GetUser(a.Ctx, msgId)

	_, _ = a.Bot.Send(tgbotapi.NewMessage(a.Cfg.Chat, string(id)))

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
