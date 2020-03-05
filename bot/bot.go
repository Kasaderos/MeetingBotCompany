package bot

import (
	"errors"
	"fmt"
	"meetingbot/google"
	"meetingbot/settings"
	"strconv"
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
	Bot      *tgbotapi.BotAPI
	mu       *sync.Mutex
	Meetings []*Meeting
	cmu      *sync.Mutex
	Chats    []*tgbotapi.Chat
	boss     int64
	Config   *settings.Config
}

func NewMeetingBot(bot *tgbotapi.BotAPI) *MeetingBot {
	return &MeetingBot{
		Bot:      bot,
		mu:       &sync.Mutex{},
		Meetings: make([]*Meeting, 0),
		Chats:    make([]*tgbotapi.Chat, 0, 10),
		cmu:      &sync.Mutex{},
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
	} else if weekdays[w.String()] == 6 {
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
		t, userNames, err := b.GetMeetTime(tMeet, events, t)
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

func (b *MeetingBot) SendInfo(chatID int64) {
	b.Bot.Send(tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("%s\n%s\n%s",
			"/daily_scrum_meeting",
			"/sprint_planing",
			"/retrospective")))
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

func (bot *MeetingBot) GetMeetTime(meetType string, events []*Event, t time.Time) (time.Time, []string, error) {
	maxMembersOfMeet := 0
	minMeetTime, maxMeetTime, duration, err := bot.GetMaxMinTime(meetType)
	if err != nil {
		return time.Time{}, []string{}, err
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
