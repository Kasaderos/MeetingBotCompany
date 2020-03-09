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
	WebhookURL = "https://altf4beta-meetingbot.herokuapp.com"
)

func main() {

	tgbot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}

	meetbot := mbot.NewMeetingBot(tgbot)
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
		panic(fmt.Errorf(err.Error()))
	}

	out := make(chan struct{})

	for update := range updates {
		text := update.Message.Text
		switch text {
		case "/start":
			{
				meetbot.AddChat(update.Message.Chat.ID)
			}
		case "/daily_scrum_meeting":
			{
				meetbot.MeetHandler(update.Message)
				meetbot.ChangeState(1, update.Message.Chat.ID)
				meetbot.ResetMoveCount(update.Message.Chat.ID)
			}
		case "/sprint_planing":
			{
				meetbot.MeetHandler(update.Message)
				meetbot.ChangeState(2, update.Message.Chat.ID)
				meetbot.ResetMoveCount(update.Message.Chat.ID)
			}
		case "/retrospective":
			{
				meetbot.MeetHandler(update.Message)
				meetbot.ChangeState(3, update.Message.Chat.ID)
				meetbot.ResetMoveCount(update.Message.Chat.ID)
			}
		case "/no":
			{
				meetbot.SendMessage("cause message:", update.Message.Chat.ID)
				switch meetbot.GetState(update.Message.Chat.ID) {
				case 1:
					meetbot.ChangeState(12, update.Message.Chat.ID)
				case 2:
					meetbot.ChangeState(22, update.Message.Chat.ID)
				case 3:
					meetbot.ChangeState(32, update.Message.Chat.ID)
				}
				meetbot.ResetMoveCount(update.Message.Chat.ID)
			}
		case "/yes":
			{
				meetbot.SendOK(update.Message.Chat.ID)
				meetbot.ChangeState(0, update.Message.Chat.ID)
				meetbot.ResetMoveCount(update.Message.Chat.ID)
			}
		case "/move":
			{
				if meetbot.GetMoveCount(update.Message.Chat.ID) == 2 {
					meetbot.SendMessage("move count == 2", update.Message.Chat.ID)
					meetbot.ChangeState(0, update.Message.Chat.ID)
					continue
				}
				meetbot.SendMessage("hh:mm-hh:mm, example:02:01-03:01", update.Message.Chat.ID)
				switch meetbot.GetState(update.Message.Chat.ID) {
				case 1:
					meetbot.ChangeState(13, update.Message.Chat.ID)
				case 2:
					meetbot.ChangeState(23, update.Message.Chat.ID)
				case 3:
					meetbot.ChangeState(33, update.Message.Chat.ID)
				}
				meetbot.IncMoveCount(update.Message.Chat.ID)
			}
		default:
			{
				switch meetbot.GetState(update.Message.Chat.ID) {
				case 12:
					meetbot.WillNotBe("daily scrum meeting", text, update.Message.Chat)
				case 22:
					meetbot.WillNotBe("sprint planing", text, update.Message.Chat)
				case 32:
					meetbot.WillNotBe("retrospective", text, update.Message.Chat)
				case 13:
					meetbot.Reshedule("daily scrum meeting", text, update.Message.Chat)
				case 23:
					meetbot.Reshedule("sprint planing", text, update.Message.Chat)
				case 33:
					meetbot.Reshedule("retrospective", text, update.Message.Chat)
				default:
					meetbot.SendInfo(update.Message.Chat.ID)
				}
			}
		}
		if strings.HasPrefix(text, "/set_alarm_") {
			msg := strings.Split(text, "_")
			go meetbot.SetNotifyTime(msg[2], update.Message.Chat.ID, out)
		} else if strings.HasPrefix(text, "/remove_alarm") {
			out <- struct{}{}
			meetbot.SendOK(update.Message.Chat.ID)
		}
	}
}
