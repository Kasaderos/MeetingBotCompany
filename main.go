package main

import (
	"fmt"
	mbot "meetingbot/bot"
	"meetingbot/settings"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const (
	BotToken   = "903632383:AAEiLoa4MKYfg5SHBEgkUNwf8AVWrxGlt8c"
	WebhookURL = "https://altf4beta-meetingbot.herokuapp.com"
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
	err = meetbot.CalcForWeek()
	if err != nil {
		fmt.Errorf(err.Error())
	}
	workTime := 5 * time.Second
	ch := make(chan struct{})
	for {
		select {
		case update := <-updates:
			cmd := StripPrefix(update.Message.Text)
			if cmd == "daily_scrum_meeting" ||
				cmd == "sprint_planing" ||
				cmd == "retrospective" {
				meetbot.Default(cmd, update.Message.Chat)
				meetbot.SendMessage(fmt.Sprintf("%s\n%s\n%s\n",
					"/will_not_be_TypeOfMeet_CauseMessage",
					"/reshedule_TypeOfMeet_hh:mm-hh:mm",
					"/will_be",
				), update.Message.Chat.ID)
			} else if strings.HasPrefix(cmd, "will_not_be_") { //will_not_be_typeMeet_I'm lazy
				msg := strings.Split(cmd, "_")
				meetbot.WillNotBe(msg[3], msg[4], update.Message.Chat)
			} else if strings.HasPrefix(cmd, "reshedule_") { // reshedule_typeMeet_12:00-15:00
				msg := strings.Split(cmd, "_")
				meetbot.Reshedule(msg[1], msg[2], update.Message.Chat)
			} else if cmd == "will_be" {
				meetbot.SendMessage("ok", update.Message.Chat.ID)
			} else if cmd == "notify_on" {
				meetbot.AddChat(update.Message.Chat)
				go Timer(ch, workTime)
			} else {
				meetbot.SendInfo(update.Message.Chat.ID)
			}
		case <-ch:
			meetbot.NotifyAll()
			go Timer(ch, workTime)
		}
	}
}

func Timer(ch chan struct{}, workTime time.Duration) {
	timer := time.NewTimer(workTime)
	select {
	case <-timer.C:
		ch <- struct{}{}
	}
}
