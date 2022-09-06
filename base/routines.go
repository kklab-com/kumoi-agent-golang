package base

import (
	"fmt"
	"sync"
	"time"

	value2 "github.com/kklab-com/goth-kkutil/value"
	kkpanic "github.com/kklab-com/goth-panic"
)

var routine = &_routine{}

const DefaultRoutineInterval = time.Second

func init() {
	go func() {
		timer := time.NewTimer(DefaultRoutineInterval)
		for {
			select {
			case now := <-timer.C:
				routine.now = now
				func() {
					defer kkpanic.Log()
					routine.timeoutTransitClean()
				}()

				timer.Reset(DefaultRoutineInterval)
			}
		}

	}()
}

type _routine struct {
	sessionPool sync.Map
	now         time.Time
}

func (r *_routine) timeoutTransitClean() {
	// maintain session and timeout request
	r.sessionPool.Range(func(key, value interface{}) bool {
		if session, ok := value.(*session); ok {
			// clean timeout request
			session.transitPool.Range(func(key, value interface{}) bool {
				tpe := value.(*transitPoolEntity)
				if tpe.timestamp.Add(DefaultTransitTimeout).Before(r.now) {
					println(fmt.Sprintf("transit timeout, %s", value2.JsonMarshal(tpe.future.SentTransitFrame())))
					session.transitPool.Delete(key)
					tpe.future.Completable().Fail(ErrTransitTimeout)
				}

				return true
			})
		}

		return true
	})
}
