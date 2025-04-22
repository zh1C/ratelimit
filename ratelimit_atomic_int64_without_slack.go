package ratelimit

import (
	"sync/atomic"
	"time"
)

type atomicInt64LimiterWithoutSlack struct {
	state      int64
	perRequest time.Duration
}

func newAtomicInt64LimiterWithoutSlack(rate int, opts ...Option) *atomicInt64LimiterWithoutSlack {
	config := buildConfig(opts)
	perRequest := config.per / time.Duration(rate)
	l := &atomicInt64LimiterWithoutSlack{
		perRequest: perRequest,
	}
	atomic.StoreInt64(&l.state, 0)
	return l
}

func (l *atomicInt64LimiterWithoutSlack) Take() time.Time {
	var (
		now                          int64
		newTimeOfNextPermissionIssue int64
	)
	for {
		now = time.Now().UnixNano()
		timeOfNextPermissionIssue := atomic.LoadInt64(&l.state)

		if timeOfNextPermissionIssue == 0 || now-timeOfNextPermissionIssue > int64(l.perRequest) {
			newTimeOfNextPermissionIssue = now
		} else {
			newTimeOfNextPermissionIssue = timeOfNextPermissionIssue + int64(l.perRequest)
		}

		if atomic.CompareAndSwapInt64(&l.state, timeOfNextPermissionIssue, newTimeOfNextPermissionIssue) {
			break
		}
	}

	sleepDuration := time.Duration(newTimeOfNextPermissionIssue - now)
	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
		return time.Unix(0, newTimeOfNextPermissionIssue)
	}

	return time.Unix(0, now)
}
