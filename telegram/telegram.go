package telegram

import (
	"bytes"
	"fmt"
	"github.com/conejoninja/home/common"
	"gopkg.in/telegram-bot-api.v4"
	"os/exec"
)

var cfg common.TelegramConfig
var bot *tgbotapi.BotAPI
var connected bool

// Start is the entrypoint
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

// Notify send a notification to every telegram client
func Notify(message string) {

	imageSrc := ""
	if cfg.RTSPSource != "" {
		cmd := exec.Command("ffmpeg", "-y -i "+cfg.RTSPSource+" -vframes 1 lastFrame.jpg")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			fmt.Print("Could not take screenshot", err)
		} else {
			imageSrc = "do.jpg"
		}
	}
	if connected {
		for _, chatID := range cfg.Chats {
			msg := tgbotapi.NewMessage(chatID, message)
			bot.Send(msg)
			if imageSrc != "" {
				msgPhoto := tgbotapi.NewPhotoUpload(chatID, "lastFrame.jpg")
				bot.Send(msgPhoto)
			}
		}
	}
}

// NotifyEvent creates and send a notification from an event
func NotifyEvent(evt common.Event) {
	if evt.Message != "" {
		msg := "⚠️ "
		if evt.Priority == 0 {
			msg = "✅ "
		}
		time := ""
		if evt.Time != nil {
			time = " [" + evt.Time.String() + "]"
		}
		Notify(msg + time + " " + evt.Message)
	}
}
