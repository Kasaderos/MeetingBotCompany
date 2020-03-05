package bot

import (
	"fmt"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func (bot *MeetingBot) Default(typeOfMeet string, chat *tgbotapi.Chat) {
	if typeOfMeet == "daily_scrum_meeting" {
		m := bot.FindMin()
		bot.SendMeet(m, chat.ID)
	} else {
		m := bot.FindMeetByType(typeOfMeet)
		if m != nil {
			bot.SendMeet(m, chat.ID)
		} else {
			bot.SendMessage("can't find meet", chat.ID)
		}
	}
}

func (m *Meeting) AddMessage(tlg, msg string) {
	for _, u := range m.Users {
		fmt.Println(tlg, u.Name)
		if userNamesTlg[tlg] == u.Name {
			u.IsWillCome = false
			u.Message = msg
			return
		}
	}
}

func (bot *MeetingBot) WillNotBe(typeOfMeet, msg string, chat *tgbotapi.Chat) {
	if typeOfMeet == "daily scrum meeting" {
		m := bot.FindMin()
		m.AddMessage(chat.UserName, msg)
		bot.SendMeet(m, chat.ID)
	} else {
		m := bot.FindMeetByType(typeOfMeet)
		if m != nil {
			m.AddMessage(chat.UserName, msg)
			bot.SendMeet(m, chat.ID)
		} else {
			bot.SendMessage("can't find meet", chat.ID)
		}
	}
	//bot.NotifyAll()
}

func (bot *MeetingBot) Reshedule(typeOfMeet, interval string, chat *tgbotapi.Chat) {

}
