package kumoi

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

func TestOmega_RemoteSession(t *testing.T) {
	ro := NewOmegaBuilder(conf).Connect().Omega()
	so := NewOmegaBuilder(conf).Connect().Omega()
	assert.NotNil(t, ro)
	assert.NotNil(t, so)
	got := false
	gc := 0

	bwg := concurrent.WaitGroup{}
	ro.GetAgentSession().OnMessage(func(msg *messages.SessionMessage) {
		println(msg.Message)
		bwg.Done()
		got = true
		gc++
	})

	srs := so.GetRemoteSession(ro.GetAgentSession().GetId())
	assert.NotNil(t, srs)
	for i := 0; i < 10; i++ {
		bwg.Add(1)
		srs.SendMessage("send")
	}

	go func() {
		<-time.After(time.Second)
		bwg.Reset()
	}()

	bwg.Wait()
	assert.True(t, got)
	assert.Equal(t, 10, gc)
	ro.Close().Await()
	so.Close().Await()
}

func TestOmega_MultiSession(t *testing.T) {
	tCount := int32(0)
	nCount := int32(0)
	bwg := concurrent.WaitGroup{}
	ng := NewOmegaBuilder(conf).Connect().Omega()
	och := ng.CreateChannel(apirequest.CreateChannel{}).Join()
	if och == nil {
		assert.Fail(t, "create ch nil")
		return
	}

	thread := 100
	times := 20
	bwg.Add(thread)
	wjd := concurrent.WaitGroup{}
	wjd.Add(thread)
	wcd := concurrent.WaitGroup{}
	wcd.Add(thread)

	go func() {
		<-time.After(time.Second * 3)
		if c := wjd.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("wjd %d", c))
		}

		wjd.Reset()
	}()

	for i := 0; i < thread; i++ {
		go func(ii int) {
			og := NewOmegaBuilder(conf).Connect().Omega()
			ch := og.GetChannel(och.Id()).Join("")
			wjd.Done()
			wjd.Wait()
			if ch == nil {
				assert.Fail(t, "get ch nil")
				return
			}

			on := int32(0)
			og.Agent().Session().OnRead(func(tf *omega.TransitFrame) {
				if tf.GetClass() == omega.TransitFrame_ClassError {
					println(tf.Error())
				}

				atomic.AddInt32(&tCount, 1)
			})

			wrd := concurrent.WaitGroup{}
			wrd.Add(thread * times)

			go func(i int) {
				<-time.After(time.Second * 30)
				if c := wrd.Remain(); c > 0 {
					assert.Fail(t, fmt.Sprintf("wrd %d", c))
					println(fmt.Sprintf("%d timeout %d", ii, on))
				}
			}(ii)

			og.Agent().OnNotification(func(tf *omega.TransitFrame) {
				atomic.AddInt32(&on, 1)
				atomic.AddInt32(&nCount, 1)
				wrd.Done()
			})

			for ir := 0; ir < times; ir++ {
				time.Sleep(time.Millisecond * 100)
				if ch.SendMessage(fmt.Sprintf("%d !!!", ir), nil) == false {
					assert.Fail(t, "send fail")
				}

				if ii == 0 {
					println(fmt.Sprintf("round %d done", ir+1))
				}
			}

			wrd.Wait()
			wcd.Done()
			wcd.Wait()
			assert.False(t, og.IsDisconnected())
			og.Close().Await()
			bwg.Done()
		}(i)
	}

	go func() {
		<-time.After(time.Second * 30)
		if c := bwg.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("bwg %d burst", c))
		}

		bwg.Reset()
	}()

	bwg.Wait()
	assert.Equal(t, int32(thread*((thread+1)*times)), tCount)
	assert.Equal(t, int32(thread*thread*times), nCount)
	println(tCount)
	println(nCount)
	och.Close()
	ng.Close().Await()
}

func TestOmega_MultiChannelCount(t *testing.T) {
	bwg := concurrent.WaitGroup{}
	ng := NewOmegaBuilder(conf).Connect().Omega()
	och := ng.CreateChannel(apirequest.CreateChannel{}).Join()
	if och == nil {
		assert.Fail(t, "create ch nil")
		return
	}

	thread := 100
	times := 100
	bwg.Add(thread)
	wjd := concurrent.WaitGroup{}
	wjd.Add(thread)
	wcd := concurrent.WaitGroup{}
	wcd.Add(thread)

	go func() {
		<-time.After(time.Second * 3)
		if c := wjd.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("wjd %d", c))
		}

		wjd.Reset()
	}()

	for i := 0; i < thread; i++ {
		go func(ii int) {
			og := NewOmegaBuilder(conf).Connect().Omega()
			ch := og.GetChannel(och.Id()).Join("")
			wjd.Done()
			wjd.Wait()
			if ch == nil {
				assert.Fail(t, "get ch nil")
				return
			}

			on := int32(0)
			og.Agent().Session().OnRead(func(tf *omega.TransitFrame) {
				if tf.GetClass() == omega.TransitFrame_ClassError {
					println(tf.Error())
				}
			})

			wrd := concurrent.WaitGroup{}
			wrd.Add(times)

			go func(i int) {
				<-time.After(time.Second * 60)
				if c := wrd.Remain(); c > 0 {
					assert.Fail(t, fmt.Sprintf("wrd %d", c))
					println(fmt.Sprintf("%d timeout %d", ii, on))
				}
			}(ii)

			og.Agent().OnResponse(func(tf *omega.TransitFrame) {
				if tf.GetChannelCount() != nil {
					atomic.AddInt32(&on, 1)
					wrd.Done()
				}
			})

			<-time.After(time.Second * 3)
			for ir := 0; ir < times; ir++ {
				time.Sleep(time.Millisecond * 100)
				assert.Equal(t, thread+1, int(ch.GetCount().Count))
				if ii == 0 {
					println(fmt.Sprintf("round %d done", ir+1))
				}
			}

			wrd.Wait()
			wcd.Done()
			wcd.Wait()
			assert.False(t, og.IsDisconnected())
			og.Close().Await()
			bwg.Done()
		}(i)
	}

	go func() {
		<-time.After(time.Second * 60)
		if c := bwg.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("bwg %d burst", c))
		}

		bwg.Reset()
	}()

	bwg.Wait()
	och.Close()
	ng.Close().Await()
}
