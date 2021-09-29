package base

import (
	"sync"
	"time"

	kkpanic "github.com/kklab-com/goth-panic"
)

var routine = &_routine{}

func init() {
	go func() {
		timer := time.NewTimer(time.Second)
		for {
			select {
			case now := <-timer.C:
				routine.now = now
				func() {
					defer kkpanic.Log()
					//routine.sessionPoolMaintainAndKeepAlive()
					routine.timeoutTransitClean()
				}()

				timer.Reset(time.Second)
			}
		}

	}()
}

type _routine struct {
	sessionPool sync.Map
	now         time.Time
}

func (r *_routine) sessionPoolMaintainAndKeepAlive() {
	// maintain session and timeout request
	r.sessionPool.Range(func(key, value interface{}) bool {
		if session, ok := value.(*Session); ok {
			if session.isClosed() {
				r.sessionPool.Delete(key)
				return true
			}

			// auto keepalive ping when no active for a while
			if session.lastActiveTimestamp.Add(DefaultKeepAlivePingInterval).Before(r.now) {
				session.ping()
			}
		}

		return true
	})
}

func (r *_routine) timeoutTransitClean() {
	// maintain session and timeout request
	r.sessionPool.Range(func(key, value interface{}) bool {
		if session, ok := value.(*Session); ok {
			// clean timeout request
			session.transitPool.Range(func(key, value interface{}) bool {
				tpe := value.(*transitPoolEntity)
				if tpe.timestamp.Add(DefaultTransitTimeout).Before(r.now) {
					session.transitPool.Delete(key)
					tpe.future.Completable().Fail(ErrTransitTimeout)
				}

				return true
			})
		}

		return true
	})
}
