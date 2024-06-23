package task

import "time"

func CallAfterDuration(d time.Duration, callback func()) func() {
	timer := time.NewTimer(d)
	cancelChan := make(chan struct{})

	go func() {
		select {
		case <-timer.C:
			callback()
		case <-cancelChan:
			timer.Stop()
		}
	}()

	return func() {
		close(cancelChan)
	}
}
