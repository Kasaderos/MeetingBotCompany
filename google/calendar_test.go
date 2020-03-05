package google

import (
	"fmt"
	"testing"
)

func TestGetEvents(t *testing.T) {
	cl, err := GetEvents("./")
	if err != nil {
		t.Errorf(err.Error())
	} else {
		for _, item := range cl.Items {
			date := item.Start.DateTime
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}
}
