package main

import (
	"fmt"
	mbot "meetingbot/bot"
	"meetingbot/settings"
	"net/http"
	"os"
	"strings"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const (
	BotToken   = "903632383:AAEiLoa4MKYfg5SHBEgkUNwf8AVWrxGlt8c"
	WebhookURL = "https://altf4beta-meeting-bot.herokuapp.com"
)

const (
	SPRINT_PLANING = iota
	DAILY_SCRUM_MEETING
	RETROSPECTIVE
)

func StripPrefix(s string) string {
	if strings.HasPrefix(s, "/") {
		return string([]byte(s)[1:])
	}
	return ""
}

func main() {
	tgbot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}
	meetbot := mbot.NewMeetingBot(tgbot)
	// bot.Debug = true
	fmt.Printf("Authorized on account %s\n", meetbot.Bot.Self.UserName)

	_, err = meetbot.Bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		panic(err)
	}
	cfg, err := settings.GetConfig()
	if err != nil {
		panic(err)
	}
	meetbot.Config = cfg
	updates := meetbot.Bot.ListenForWebhook("/")

	port := os.Getenv("PORT")
	go http.ListenAndServe(":"+port, nil)
	fmt.Println("start listen :8080")
	err := meetbot.CalcForWeek()
	if err != nil {
		fmt.Errorf(err.Error())
	}
	for update := range updates {
		cmd := StripPrefix(update.Message.Text)
		if cmd == "daily_scrum_meeting" ||
			cmd == "sprint_planing" ||
			cmd == "retrospective" {
			meetbot.Default(cmd, update.Message.Chat)
		} else {
			meetbot.SendInfo(update.Message.Chat.ID)
		}
	}
}
