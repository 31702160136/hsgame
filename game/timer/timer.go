package timer

import (
	"fmt"
	"game/dispatch"
	"game/log"
	"runtime/debug"
	"sync"
	"time"
)

type timer struct {
	*time.Timer
	Close chan byte
}

var (
	timers    = make(map[string]*timer)
	timersMux = &sync.Mutex{}
)

func newTimer(sec time.Duration) *timer {
	return &timer{
		Timer: time.NewTimer(sec),
		Close: make(chan byte, 1),
	}
}

func After(name string, sec time.Duration, cbFunc interface{}, args ...interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("timer: %s ==> %v\n%s", name, err, string(debug.Stack()))
			}
		}()

		timersMux.Lock()
		t := newTimer(sec)
		timers[name] = t
		timersMux.Unlock()
		select {
		case <-t.C:
			dispatch.PushSystemSyncMsg(fmt.Sprintf("after:%s", name), cbFunc, args...)
		case <-t.Close:
		}
	}()
}

func Stop(name string) {
	timersMux.Lock()
	t, ok := timers[name]
	timersMux.Unlock()
	if !ok {
		return
	}
	t.Stop()
	delete(timers, name)
}

func AfterGo(name string, sec time.Duration, cbFunc interface{}, args ...interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("timer: %s ==> %v\n%s", name, err, string(debug.Stack()))
			}
		}()

		timersMux.Lock()
		t := newTimer(sec)
		timers[name] = t
		timersMux.Unlock()
		select {
		case <-t.C:
			dispatch.PushSystemGoMsg(fmt.Sprintf("afterGo:%s", name), cbFunc, args...)
		case <-t.Close:
		}
	}()
}

func (this *timer) Stop() {
	this.Close <- 1
	this.Timer.Stop()
}
