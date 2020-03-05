package bot

import (
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

func AddMessage(m *Meeting, tlg string) {
	for _, u := range m.Users {
		if userNamesTlg[chat.UserName] == u.Name {
			u.Message = msg
			return
		}
	}
}

func (bot *MeetingBot) WillNotBe(typeOfMeet, msg string, chat *tgbotapi.Chat) {
	if typeOfMeet == "daily_scrum_meeting" {
		m := bot.FindMin()
		AddMessage(m, chat.UserName)
	} else {
		m := bot.FindMeetByType(typeOfMeet)
		if m != nil {
			bot.SendMeet(m, chat.ID)
			AddMessage(m, chat.UserName)
		} else {
			bot.SendMessage("can't find meet", chat.ID)
		}
	}
	bot.NotifyAll()
}

func (bot *MeetingBot) Reshedule(typeOfMeet, interval string, chat *tgbotapi.Chat) {

}
