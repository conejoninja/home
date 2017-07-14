package telegram

import (
	"fmt"

	"github.com/conejoninja/home/common"
	"gopkg.in/telegram-bot-api.v4"
)

var cfg common.TelegramConfig
var bot *tgbotapi.BotAPI
var connected bool

func Start(homecfg common.HomeConfig) {
	cfg = homecfg.Tg

	var err error
	bot, err = tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		fmt.Println("Error starting Telegram bot:", err)
	} else {
		connected = true
	}
	bot.Debug = false
}

func Notify(message string) {
	if connected {
		for _, chatID := range cfg.Chats {
			msg := tgbotapi.NewMessage(chatID, message)
			bot.Send(msg)
		}
	}
}

func NotifyEvent(evt common.Event) {
	if evt.Message != "" {
		msg := "⚠️ "
		if evt.Priority == 0 {
			msg = "✅ "
		}
		Notify(msg + evt.Message)
	}
}
