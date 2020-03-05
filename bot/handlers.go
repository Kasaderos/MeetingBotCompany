package bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func (bot *MeetingBot) FindMeetByType(t string) *Meeting {
	for i, v := range bot.Meetings {
		if v.Type == t {
			return v
		}
	}
	return nil
}
func (bot *MeetingBot) FindMin() *Meeting {
	min := time.Now().AddDate(10, 0, 0)
	ind := 0
	t := time.Time{}
	for i, v := range bot.Meetings {
		fmt.Println(v.Type)
		fmt.Println(v.Date)
		for _, vv := range v.Users {
			fmt.Println(vv.Name)
		}
		fmt.Println()
		if v.Date.Before(min) && v.Date != t {
			min = v.Date
			ind = i
		}
	}
	return bot.Meetings[ind]
}

var (
	userNames = map[string]string{
		"kberda99@gmail.com":          "Kaldarov Berdibek",
		"rayskiy.vladimirr@gmail.com": "Vladimir Savostin",
		"aidar.babanov@nu.edu.kz":     "Aidar Babanov",
	}
)

func GetUsersFromSlice(names []string) []*User {
	users := make([]*User, 0, len(names))
	for _, v := range names {
		users = append(users, &User{
			Name: userNames[v],
		})
	}
	return users
}

func GetNamesUsers(users []*User) string {
	names := make([]string, 0, len(users))
	for _, u := range users {
		names = append(names, u.Name)
	}
	return strings.Join(names, "\n")
}

func (bot *MeetingBot) SendMeet(m *Meeting, chatID int64) {
	bot.Bot.Send(tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("%s\n%s\n%s\n",
			m.Type,
			m.Date.Format(time.UnixDate),
			GetNamesUsers(m.Users))))
}

func (bot *MeetingBot) SendMessage(msg string, chatID int64) {
	bot.Bot.Send(tgbotapi.NewMessage(
		chatID,
		msg,
	))
}

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
