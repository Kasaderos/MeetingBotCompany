package bot

import (
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func (bot *MeetingBot) Default(typeOfMeet string, chat *tgbotapi.Chat) {
	bot.AddChat(chat)
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
}

func (bot *MeetingBot) Reshedule(typeOfMeet, interval string, chat *tgbotapi.Chat) {
	if typeOfMeet == "daily scrum meeting" {
		m := bot.FindMin()
		logs.Debug("daily FindMin")
		m = bot.Recalc(typeOfMeet, interval, m, chat.ID)
		logs.Debug("After Calc")
		bot.SendMeet(m, chat.ID)
	} else {
		logs.Debug("others")
		m := bot.FindMeetByType(typeOfMeet)
		if m != nil {
			logs.Debug("before recalc")
			m = bot.Recalc(typeOfMeet, interval, m, chat.ID)
			logs.Debug("after recalc")
			bot.SendMeet(m, chat.ID)
		} else {
			bot.SendMessage("can't find meet", chat.ID)
		}
	}
}

func (bot *MeetingBot) Recalc(typeOfMeet, interval string, m *Meeting, chatID int64) *Meeting {
	minMaxTime := strings.Split(interval, "-")
	_, _, d, err := bot.GetMaxMinTime(typeOfMeet)
	if err != nil {
		bot.SendMessage("error: get duration", chatID)
	}
	minMeetTime, err1 := time.Parse(time.RFC3339, "2006-01-02T"+minMaxTime[0]+":00+06:00")
	maxMeetTime, err2 := time.Parse(time.RFC3339, "2006-01-02T"+minMaxTime[1]+":00+06:00")
	if err1 != nil || err2 != nil {
		bot.SendMessage("error: Parse interval", chatID)
	}

	newTime, userNames, err := bot.GetMeetTime(typeOfMeet, m.Events, m.Date, minMeetTime, maxMeetTime, d)
	if err != nil {
		bot.SendMessage("error: Recalc next meet", chatID)
	}
	users := GetUsersFromSlice(userNames)
	m.Users = users
	m.Date = newTime
	return m
}
