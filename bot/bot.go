package bot

import (
	"errors"
	"fmt"
	"meetingbot/google"
	"meetingbot/settings"
	"strconv"
	"strings"
	"sync"
	"time"

	calendar "google.golang.org/api/calendar/v3"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var (
	weekdays = map[string]int{
		"Monday":    1,
		"Tuesday":   2,
		"Wednesday": 3,
		"Thursday":  4,
		"Friday":    5,
		"Saturday":  6,
		"Sunday":    7,
	}
)

type MeetingBot struct {
	Bot        *tgbotapi.BotAPI
	mu         *sync.Mutex
	Meetings   []*Meeting
	cmu        *sync.Mutex
	Chats      []*tgbotapi.Chat
	boss       int64
	Config     *settings.Config
	NotifyTime time.Duration
}

func NewMeetingBot(bot *tgbotapi.BotAPI) *MeetingBot {
	return &MeetingBot{
		Bot:        bot,
		mu:         &sync.Mutex{},
		Meetings:   make([]*Meeting, 0),
		Chats:      make([]*tgbotapi.Chat, 0, 10),
		cmu:        &sync.Mutex{},
		NotifyTime: time.Second * 5,
	}
}

type Event struct {
	Creator string
	Start   time.Time
	End     time.Time
}

type User struct {
	Name       string
	IsWillCome bool
	Message    string
}

type Meeting struct {
	Type   string
	Date   time.Time
	Users  []*User
	Events []*Event
}

func GetMeetType(t time.Time) string {
	w := t.Weekday()
	if weekdays[w.String()] == 1 {
		return "sprint planing"
	} else if weekdays[w.String()] == 5 {
		return "retrospective"
	}
	return "daily scrum meeting"
}

func (b *MeetingBot) CalcForWeek() error {
	calendarEvents, err := b.GetEvents()
	if err != nil {
		return err
	}
	meetings := make([]*Meeting, 0, 7)
	for i := 0; i < 7; i++ {
		t := time.Now().AddDate(0, 0, i)
		fmt.Println("time", t.String())
		events := make([]*Event, 0)
		for _, item := range calendarEvents.Items {
			start := item.Start.DateTime
			end := item.End.DateTime
			s, err1 := time.Parse(time.RFC3339, start)
			e, err2 := time.Parse(time.RFC3339, end)
			if err1 != nil || err2 != nil {
				return errors.New("parse")
			}
			y, m, d := t.Date()
			y1, m1, d1 := s.Date()
			if y == y1 && m == m1 && d == d1 {
				events = append(events, &Event{
					Creator: item.Creator.Email,
					Start:   s,
					End:     e,
				})
			}
		}
		tMeet := GetMeetType(t)
		minMeetTime, maxMeetTime, duration, err := b.GetMaxMinTime(tMeet)
		if err != nil {
			return err
		}
		t, userNames, err := b.GetMeetTime(tMeet, events, t, minMeetTime, maxMeetTime, duration)
		if err != nil {
			return err
		}
		users := GetUsersFromSlice(userNames)
		m := &Meeting{
			Type:   tMeet,
			Date:   t,
			Events: events,
			Users:  users,
		}
		meetings = append(meetings, m)
	}
	b.Meetings = meetings
	return nil
}

func (b *MeetingBot) NotifyAll() {
	for _, chat := range b.Chats {
		b.Default("daily_scrum_meeting", chat)
	}
}

func (b *MeetingBot) SendInfo(chatID int64) {
	b.Bot.Send(tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("%s\n%s\n%s\n%s",
			"/daily_scrum_meeting",
			"/sprint_planing",
			"/retrospective",
			"/notify_on")))
}

// всем посылает сообщения
func (b *MeetingBot) GetMaxMinTime(t string) (time.Time, time.Time, time.Duration, error) {
	for _, m := range b.Config.Meetings {
		if m.Type == t {
			minMeetTime, err1 := time.Parse(time.RFC3339, "2006-01-02T"+m.MinStart+":00+06:00")
			maxMeetTime, err2 := time.Parse(time.RFC3339, "2006-01-02T"+m.MaxStart+":00+06:00")
			if err1 != nil || err2 != nil {
				return time.Time{}, time.Time{}, time.Second, errors.New("GetMaxMinTime " + m.MinStart + " " + m.MaxStart)
			}

			return minMeetTime, maxMeetTime, time.Minute * time.Duration(m.Duration), nil
		}
	}
	return time.Time{}, time.Time{}, time.Second, errors.New("GetMaxMinTime")
}

func (bot *MeetingBot) GetMeetTime(meetType string, events []*Event, t time.Time,
	minMeetTime, maxMeetTime time.Time,
	duration time.Duration) (time.Time, []string, error) {

	maxMembersOfMeet := 0
	if len(events) == 0 {
		return time.Time{}, []string{}, nil
	}
	minTime := minMeetTime
	maxTime := minMeetTime.Add(duration)
	var meetTime time.Time
	membersName := make([]string, 0)
	for true {
		minHH, minMM, _ := minTime.Clock()
		maxHH, maxMM, _ := maxTime.Clock()
		mHH, mMM, _ := maxMeetTime.Clock()
		if (mHH*60 + mMM) < (maxHH*60 + maxMM) {
			break
		}
		count := 0
		tempMembersName := make([]string, 0)
		for _, ev := range events {
			hStart, mStart, _ := ev.Start.Clock()
			hEnd, mEnd, _ := ev.End.Clock()
			if ((minHH*60 + minMM) >= (hStart*60.0 + mStart)) && ((maxHH*60 + maxMM) <= (hEnd*60.0 + mEnd)) {
				count++
				tempMembersName = append(tempMembersName, ev.Creator)
			}
		}
		if maxMembersOfMeet < count {
			maxMembersOfMeet = count
			h, m, s := t.Clock()
			dur1, err := ParseDuration(h, m, s)
			if err != nil {
				return time.Time{}, []string{}, err
			}
			h, m, s = minTime.Clock()
			dur2, err := ParseDuration(h, m, s)
			if err != nil {
				return time.Time{}, []string{}, err
			}
			meetTime = t.Add(-dur1).Add(dur2)
			membersName = tempMembersName
		}
		minTime = minTime.Add(time.Minute)
		maxTime = maxTime.Add(time.Minute)
	}
	return meetTime, membersName, nil
}

// получает даты с календаря
func (bot *MeetingBot) GetEvents() (*calendar.Events, error) {
	events, err := google.GetEvents("google/")
	if err != nil {
		return nil, err
	}
	return events, nil
}

func ParseDuration(h, m, s int) (time.Duration, error) {
	dur, err := time.ParseDuration(
		strconv.Itoa(h) + "h" +
			strconv.Itoa(m) + "m" +
			strconv.Itoa(s) + "s")
	if err != nil {
		return time.Duration(0), err
	}
	return dur, nil
}
func (b *MeetingBot) AddChat(chat *tgbotapi.Chat) {
	b.cmu.Lock()
	for _, v := range b.Chats {
		if v.ID == chat.ID {
			b.cmu.Unlock()
			return
		}
	}
	b.Chats = append(b.Chats, chat)
	b.cmu.Unlock()
}

func (b *MeetingBot) DeleteChat(chatID int64) error {
	b.cmu.Lock()
	for i, v := range b.Chats {
		if v.ID == chatID {
			if len(b.Chats) > 1 {
				b.Chats[i], b.Chats[len(b.Chats)-1] = b.Chats[len(b.Chats)-1], b.Chats[i]
				b.Chats = b.Chats[:len(b.Chats)-1]
			} else {
				b.Chats = make([]*tgbotapi.Chat, 0)
			}
			b.cmu.Unlock()
			return nil
		}
	}
	b.cmu.Unlock()
	return errors.New("not found chatID")
}

func (b *MeetingBot) SendOK(chatID int64) {
	b.Bot.Send(tgbotapi.NewMessage(
		chatID,
		"ok"))
}

func (bot *MeetingBot) FindMeetByType(t string) *Meeting {
	t = strings.ReplaceAll(t, "_", " ")
	for _, v := range bot.Meetings {
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
	userNamesTlg = map[string]string{
		"berda0_o":          "Kaldarov Berdibek",
		"Savostin_Vladimir": "Vladimir Savostin",
	}
)

func GetUsersFromSlice(names []string) []*User {
	users := make([]*User, 0, len(names))
	for _, v := range names {
		users = append(users, &User{
			Name:       userNames[v],
			IsWillCome: true,
		})
	}
	return users
}

func GetInfoUsers(users []*User) string {
	names := make([]string, 0, len(users))
	for _, u := range users {
		if !u.IsWillCome {
			names = append(names, u.Name+" [NO] "+u.Message)
		} else {
			names = append(names, u.Name+"[YES]")
		}
	}
	return strings.Join(names, "\n")
}

func (bot *MeetingBot) SendMeet(m *Meeting, chatID int64) {
	bot.Bot.Send(tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("%s\n%s\n%s\n",
			m.Type,
			m.Date.Format(time.UnixDate),
			GetInfoUsers(m.Users))))
}

func (bot *MeetingBot) SendMessage(msg string, chatID int64) {
	bot.Bot.Send(tgbotapi.NewMessage(
		chatID,
		msg,
	))
}

type Alarm struct {
	Hours   int
	Minutes int
	Timeout chan time.Time
}

func (alarm *Alarm) SetAlarm(Hours, Minutes int) {

}
func (bot *MeetingBot) SetNotifyTime(t string, chatID int64) {
	clock := strings.Split(t, ":")
	h, err1 := strconv.Atoi(clock[0])
	m, err2 := strconv.Atoi(clock[1])
	if err1 != nil || err2 != nil {
		bot.SendMessage("error: strconv", chatID)
	}
	hh, mm, ss := time.Now().Clock()
	tomorrow := time.Now().AddDate(0, 0, 1).Add(-time.Second * time.Duration(hh*3600+mm*60+ss))
	tomorrow.Add(time.Hour * time.Duration(h))
	tomorrow.Add(time.Minute * time.Duration(m))
}
