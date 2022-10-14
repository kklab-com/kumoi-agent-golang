package kumoi

import (
	"fmt"
	"math"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

func TestChannel_ChannelJoin(t *testing.T) {
	omg := NewOmegaBuilder(conf).Connect().Get()
	chf := omg.CreateChannel(apirequest.CreateChannel{
		Name:              t.Name(),
		IdleTimeoutSecond: 100,
	})

	assert.NotNil(t, chf.Get())
	chInfo := chf.Info()
	assert.NotNil(t, chInfo)
	assert.Equal(t, t.Name(), chInfo.Name())
	ch := chInfo.Join(chf.Get().ParticipatorKey)
	assert.NotNil(t, ch)
	vf := concurrent.NewFuture()
	ch.OnLeave(func() {
		vf.Completable().Complete(nil)
	})

	cmMeta := concurrent.NewFuture()
	ch.Watch(func(msg messages.TransitFrame) {
		println(reflect.ValueOf(msg).Elem().Type().Name())
		switch v := msg.(type) {
		case *messages.ChannelMessage:
			if v.GetMetadata() != nil && v.GetMetadata()["typ"] == "SendMessage" {
				cmMeta.Completable().Complete(nil)
			}
		}
	})

	assert.NotNil(t, ch)
	assert.NotEqual(t, ch.Role(), omega.Role_RoleOwner)
	assert.True(t, ch.SendMessage("SendMessage", map[string]interface{}{"typ": "SendMessage"}).AwaitTimeout(Timeout).IsSuccess())
	assert.False(t, ch.SendOwnerMessage("SendOwnerMessage", nil).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, cmMeta.Await().IsSuccess())
	lm := ch.Leave().Await().TransitFrame()
	assert.NotEmpty(t, lm.GetChannelId())
	assert.NotNil(t, chInfo.Join(""))

	assert.False(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.Nil(t, chInfo.Join(""))
	assert.True(t, ch.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vf.AwaitTimeout(Timeout).IsSuccess())

	ch = chf.Join()
	closed := false
	ch.OnClose(func() {
		closed = true
	})

	ch.Watch(func(msg messages.TransitFrame) {
		println(reflect.ValueOf(msg).Elem().Type().Name())
	})

	assert.NotNil(t, ch)
	assert.Equal(t, ch.Role(), omega.Role_RoleOwner)
	assert.True(t, ch.SendOwnerMessage("SendOwnerMessage", nil).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, ch.SetName("new_channel_name").AwaitTimeout(Timeout).IsSuccess())
	println("wait for replay")
	<-time.After(2 * time.Second)
	cp := ch.ReplayChannelMessage(0, true, omega.Volume_VolumeLowest)
	assert.NotNil(t, cp)
	assert.Equal(t, int32(1), ch.Count().TransitFrame().GetCount())
	assert.Equal(t, ch.Id(), cp.Next().Cast().GetChannelMeta().GetChannelId())
	assert.Equal(t, ch.Name(), cp.Next().Cast().SetChannelMeta().GetName())
	assert.Equal(t, string((&omega.ChannelOwnerMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, string((&omega.ChannelMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Nil(t, cp.Next())
	cm := ch.Close().AwaitTimeout(Timeout)
	assert.True(t, cm.IsSuccess())
	assert.NotEmpty(t, cm.TransitFrame().GetChannelId())
	println("wait for playback")
	<-time.After(2 * time.Second)
	cp = omg.PlaybackChannelMessage(ch.Id(), math.MaxInt32, false, omega.Volume_VolumeLowest)
	assert.NotNil(t, cp)
	assert.Equal(t, string((&omega.ChannelMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, string((&omega.ChannelOwnerMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, ch.Name(), cp.Next().Cast().SetChannelMeta().GetName())
	assert.Equal(t, ch.Id(), cp.Next().Cast().GetChannelMeta().GetChannelId())
	assert.Equal(t, ch.Id(), cp.Next().Cast().CloseChannel().GetChannelId())
	assert.Nil(t, cp.Next())
	assert.True(t, closed)
	assert.True(t, omg.Close().Await().AwaitTimeout(Timeout).IsSuccess())
}

func TestOmega_ChannelMultiSession(t *testing.T) {
	nCount := int32(0)
	connectWG := concurrent.WaitGroup{}
	omg := NewOmegaBuilder(conf).Connect().Get()
	och := omg.CreateChannel(apirequest.CreateChannel{}).Join()
	if och == nil {
		assert.Fail(t, "create ch nil")
		return
	}

	defer func() {
		och.Close().Await()
		omg.Close().Await()
	}()

	thread := 50
	times := 20

	connectWG.Add(thread)
	joinWG := concurrent.WaitGroup{}
	joinWG.Add(thread)
	routineWG := concurrent.WaitGroup{}
	routineWG.Add(thread)

	go func() {
		<-time.After(time.Second * 5)
		if c := joinWG.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("joinWG %d", c))
		}

		joinWG.Reset()
	}()

	for i := 0; i < thread; i++ {
		go func(threadIndex int) {
			tog := NewOmegaBuilder(conf).Connect().Get()
			ch := tog.Channel(och.Id()).Join("")
			joinWG.Done()
			joinWG.Wait()
			if ch == nil {
				assert.Fail(t, "get ch nil")
				return
			}

			on := int32(0)
			wrd := concurrent.WaitGroup{}
			wrd.Add(thread * times)

			go func(i int) {
				<-time.After(time.Second * 30)
				if c := wrd.Remain(); c > 0 {
					assert.Fail(t, fmt.Sprintf("wrd %d", c))
					println(fmt.Sprintf("%d timeout %d", threadIndex, on))
				}
			}(threadIndex)

			ch.Watch(func(msg messages.TransitFrame) {
				if msg.Cast().ChannelMessage() != nil {
					atomic.AddInt32(&on, 1)
					atomic.AddInt32(&nCount, 1)
					wrd.Done()
				}
			})

			for ir := 0; ir < times; ir++ {
				time.Sleep(time.Millisecond * 100)
				if !ch.SendMessage(fmt.Sprintf("%d !!!", ir), nil).AwaitTimeout(Timeout * 10).IsSuccess() {
					assert.Fail(t, "send fail")
				}

				if threadIndex == 0 {
					println(fmt.Sprintf("round %d done", ir+1))
				}
			}

			wrd.Wait()
			routineWG.Done()
			routineWG.Wait()
			tog.Close().Await()
			connectWG.Done()
		}(i)
	}

	go func() {
		<-time.After(time.Second * 30)
		if c := connectWG.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("connectWG %d burst", c))
		}

		connectWG.Reset()
	}()

	connectWG.Wait()
	assert.Equal(t, int32(thread*thread*times), nCount)
	println(nCount)
}

func TestOmega_ChannelMultiChannelCount(t *testing.T) {
	bwg := concurrent.WaitGroup{}
	omg := NewOmegaBuilder(conf).Connect().Get()
	och := omg.CreateChannel(apirequest.CreateChannel{}).Join()
	if och == nil {
		assert.Fail(t, "create ch nil")
		return
	}

	thread := 50
	times := 50
	bwg.Add(thread)
	wjd := concurrent.WaitGroup{}
	wjd.Add(thread)
	wcd := concurrent.WaitGroup{}
	wcd.Add(thread)

	go func() {
		<-time.After(time.Second * 10)
		if c := wjd.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("fail all connect, wjd %d", c))
		}

		wjd.Reset()
	}()

	for i := 0; i < thread; i++ {
		go func(ii int) {
			og := NewOmegaBuilder(conf).Connect().Get()
			ch := og.Channel(och.Id()).Join("")
			println(fmt.Sprintf("%s connected", og.Session().GetId()))
			wjd.Done()
			wjd.Wait()
			println("all connected, go")
			if ch == nil {
				assert.Fail(t, "get ch nil")
				return
			}

			on := int32(0)
			wrd := concurrent.WaitGroup{}
			wrd.Add(times)

			go func(i int) {
				<-time.After(time.Second * 30)
				if c := wrd.Remain(); c > 0 {
					assert.Fail(t, fmt.Sprintf("wrd %d", c))
					println(fmt.Sprintf("%d timeout %d", i, on))
				}
			}(ii)

			og.OnMessageHandler = func(tf messages.TransitFrame) {
				if tf.Cast().ChannelCount() != nil {
					atomic.AddInt32(&on, 1)
					wrd.Done()
				}
			}

			<-time.After(time.Second * 3)
			for ir := 0; ir < times; ir++ {
				assert.Equal(t, thread+1, int(ch.Count().TransitFrame().GetCount()))
				if ii == 0 {
					println(fmt.Sprintf("round %d done", ir+1))
				}
			}

			wrd.Wait()
			wcd.Done()
			wcd.Wait()
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
	omg.Close().Await()
}
