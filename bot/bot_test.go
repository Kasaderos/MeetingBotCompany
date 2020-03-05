package bot

import (
	"fmt"
	"testing"
	"time"

	calendar "google.golang.org/api/calendar/v3"
)

type IntersectAns struct {
	ans   string
	empty bool
}

func TestGetMeetTime(t *testing.T) {
	events := &calendar.Events{
		Items: []*calendar.Event{
			&calendar.Event{
				Creator: &calendar.EventCreator{
					DisplayName: "Mike",
				},
				Start: &calendar.EventDateTime{
					DateTime: "2006-01-02T10:00:00+06:00",
				},
				End: &calendar.EventDateTime{
					DateTime: "2006-01-02T18:00:00+06:00",
				},
			},
			&calendar.Event{
				Creator: &calendar.EventCreator{
					DisplayName: "Mike2",
				},
				Start: &calendar.EventDateTime{
					DateTime: "2006-01-02T14:00:00+06:00",
				},
				End: &calendar.EventDateTime{
					DateTime: "2006-01-02T18:00:00+06:00",
				},
			},
			&calendar.Event{
				Creator: &calendar.EventCreator{
					DisplayName: "Mike3",
				},
				Start: &calendar.EventDateTime{
					DateTime: "2006-01-02T10:00:00+06:00",
				},
				End: &calendar.EventDateTime{
					DateTime: "2006-01-02T18:00:00+06:00",
				},
			},
		},
	}
	T, _ := time.Parse(time.RFC3339, "2006-01-02T10:00:00+06:00")
	bot := MeetingBot{}
	m, err := bot.GetMeetTime("daily", events, T, nil, nil)
	fmt.Println(m)
	if err != nil {
		t.Error(err.Error())
	}
}
