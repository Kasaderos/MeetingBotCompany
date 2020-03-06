package timer

import (
	"time"
)

func SetTimer(ch chan struct{}, workTime time.Duration) {
	timer := time.NewTimer(workTime)
	select {
	case <-timer.C:
		ch <- struct{}{}
	}
}