package sotatgbot

import (
	"context"
	"fmt"

	"github.com/CookieNyanCloud/tgFeedBackBot/configs"
	"github.com/CookieNyanCloud/tgFeedBackBot/repository/database/redisDB"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	redErr  = "error in redis:%v\n"
	welcome = "Привет, я связующая бездна"
	Next    = "Вперед"
)

type Actions struct {
	Ctx   context.Context
	Cache *redisDB.Repo
	Bot   *tgbotapi.BotAPI
	Cfg   *configs.Conf
}

func NewActions(
	ctx context.Context,
	cache *redisDB.Repo,
	bot *tgbotapi.BotAPI,
	cfg *configs.Conf) *Actions {
	return &Actions{
		Cache: cache,
		Bot:   bot,
		Ctx:   ctx,
		Cfg:   cfg,
	}
}

type ActionsInterface interface {
	ReplyToMsgTxt(chatId int, txt string)
	ReplyToMsgTxtById(id int64, txt string)
	ReplyToMsgMedia(chatId int, mediaID string, tgType string)
	ReplyToMsgPhotoVideo(chatId int, mediaID string, tgType, text string)
	ReplyToMsgFile(chatId int, txt string)
	SendMsg(chatId int64, msgId int)
	BanUser(msgId int)
	CheckBanUser(chatId int64) bool
}

func (a *Actions) StartMsg(chatId int64) {
	msg := tgbotapi.NewMessage(chatId, welcome)
	_, _ = a.Bot.Send(msg)

}

func (a *Actions) ReplyToMsgTxtById(id int64, txt string) {
	msg := tgbotapi.NewMessage(id, txt)
	_, _ = a.Bot.Send(msg)
}

func (a *Actions) ReplyToMsgTxt(chatId int, txt string) {
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

func (a *Actions) ReplyToMsgMedia(chatId int, mediaID, tgType string) {
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
	media := tgbotapi.FileID(mediaID)
	var msg tgbotapi.Chattable
	switch tgType {
	case "sticker":
		msg = tgbotapi.NewSticker(id, media)
	case "voice":
		msg = tgbotapi.NewVoice(id, media)
	}
	_, _ = a.Bot.Send(msg)

}

func (a *Actions) ReplyToMsgPhotoVideo(chatId int, mediaID, tgType, text string) {
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
	media := tgbotapi.FileID(mediaID)
	switch tgType {
	case "photo":
		msg := tgbotapi.NewPhoto(id, media)
		msg.Caption = text
		_, _ = a.Bot.Send(msg)
	case "video":
		msg := tgbotapi.NewVideo(id, media)
		msg.Caption = text
		_, _ = a.Bot.Send(msg)
	}
}

func (a *Actions) ReplyToMsgFile(chatId int, fileID string) {
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
	file := tgbotapi.RequestFileData(tgbotapi.FileID(fileID))
	msg := tgbotapi.NewDocument(id, file)
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
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
		RemoveKeyboard: false,
		Selective:      false,
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
	if err != nil && err != redis.Nil {
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
