package ratelimit

import (
	"sync"
	"time"
)

type mutexLimiterWithoutSlack struct {
	sync.Mutex
	last       time.Time
	perRequest time.Duration
	sleepFor   time.Duration
}

func newMutexLimiterWithoutSlack(rate int, opts ...Option) *mutexLimiterWithoutSlack {
	config := buildConfig(opts)
	perRequest := config.per / time.Duration(rate)
	l := &mutexLimiterWithoutSlack{
		perRequest: perRequest,
	}
	return l
}

func (l *mutexLimiterWithoutSlack) Take() time.Time {
	l.Lock()
	defer l.Unlock()

	now := time.Now()
	if l.last.IsZero() {
		l.last = now
		return l.last
	}

	l.sleepFor += l.perRequest - now.Sub(l.last)

	if l.sleepFor > 0 {
		time.Sleep(l.sleepFor)
		l.last = now.Add(l.sleepFor)
		l.sleepFor = 0
	} else {
		l.last = now
	}

	return l.last
}
