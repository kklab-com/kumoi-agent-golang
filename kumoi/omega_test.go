package kumoi

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

var appId = ""
var token = ""
var domain = ""
var conf *base.Config

const Timeout = time.Second
const TestMessage = "MESSAGE"

func TestMain(m *testing.M) {
	appId = os.Getenv("TEST_APP_ID")
	token = os.Getenv("TEST_TOKEN")
	domain = os.Getenv("TEST_DOMAIN")
	conf = base.NewConfig(appId, token)
	conf.Domain = domain
	println(fmt.Sprintf("domain %s", domain))
	m.Run()
}

func TestOmega(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Get()
	assert.NotEmpty(t, o)
	assert.NotEmpty(t, o.Hello())
	assert.NotEmpty(t, o.ServerTime())
	assert.True(t, o.Broadcast("golang broadcast test"))
	chf := o.CreateChannel(apirequest.CreateChannel{
		Name:              "!!!",
		IdleTimeoutSecond: 300,
	})

	assert.NotNil(t, chf.Get())
	chInfo := chf.Info()
	assert.NotNil(t, chInfo)
	ch := chInfo.Join(chf.Get().ParticipatorKey)
	vfl := concurrent.NewFuture()
	ch.OnLeave(func() {
		vfl.Completable().Complete(nil)
	})

	ch.Watch(func(msg messages.TransitFrame) {
		println(reflect.ValueOf(msg).Elem().Type().Name())
	})

	assert.NotNil(t, ch)
	assert.NotEqual(t, ch.Role(), omega.Role_RoleOwner)
	assert.True(t, ch.SendMessage("SendMessage", nil).AwaitTimeout(Timeout).IsSuccess())
	assert.False(t, ch.SendOwnerMessage("SendOwnerMessage", nil).AwaitTimeout(Timeout).IsSuccess())
	assert.False(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.Nil(t, chInfo.Join(""))
	assert.True(t, ch.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vfl.AwaitTimeout(Timeout).IsSuccess())

	ch = chf.Join()
	vfc := concurrent.NewFuture()
	ch.OnClose(func() {
		vfc.Completable().Complete(nil)
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
	assert.Equal(t, ch.Id(), cp.Next().(*messages.GetChannelMeta).GetChannelId())
	assert.Equal(t, ch.Name(), cp.Next().(*messages.SetChannelMeta).GetName())
	assert.Equal(t, string((&omega.ChannelOwnerMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, string((&omega.ChannelMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Nil(t, cp.Next())
	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())
	println("wait for playback")
	<-time.After(2 * time.Second)
	cp = o.PlaybackChannelMessage(ch.Id(), math.MaxInt32, false, omega.Volume_VolumeLowest)
	assert.NotNil(t, cp)
	assert.Equal(t, string((&omega.ChannelMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, string((&omega.ChannelOwnerMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, ch.Name(), cp.Next().(*messages.SetChannelMeta).GetName())
	assert.Equal(t, ch.Id(), cp.Next().(*messages.GetChannelMeta).GetChannelId())
	assert.Equal(t, ch.Id(), cp.Next().(*messages.CloseChannel).GetChannelId())
	assert.Nil(t, cp.Next())
	assert.True(t, vfc.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, o.Close().Await().AwaitTimeout(Timeout).IsSuccess())
}

func TestOmegaClose(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Get()
	bwg := concurrent.WaitGroup{}
	go func() {
		<-time.After(3 * time.Second)
		if bwg.Remain() > 0 {
			bwg.Reset()
			assert.Fail(t, "disconnected not invoke")
		}
	}()

	bwg.Add(1)
	o.OnClosedHandler = func() {
		bwg.Done()
	}

	go func() {
		assert.False(t, o.IsClosed())
		<-time.After(1 * time.Second)
		o.Close().Await()
		assert.True(t, o.IsClosed())
	}()

	bwg.Wait()
	assert.True(t, o.IsClosed())
}

func TestOmegaWriteOnClosed(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Get()
	chResp := o.CreateChannel(apirequest.CreateChannel{})
	chInfo := chResp.Info()
	ch := chInfo.Join("")
	assert.True(t, ch.SendMessage("!!!", nil).AwaitTimeout(Timeout).IsSuccess())
	o.Close().Await()
	assert.False(t, ch.SendMessage("!!!", nil).AwaitTimeout(Timeout).IsSuccess())
}

func TestOmega_MultiVoteChannel(t *testing.T) {
	nCount := int32(0)
	bwg := concurrent.WaitGroup{}
	ng := NewOmegaBuilder(conf).Connect().Get()
	och := ng.CreateChannel(apirequest.CreateChannel{Name: fmt.Sprintf("%s_C", t.Name())}).Join()
	if och == nil {
		assert.Fail(t, "create ch nil")
		return
	}

	ovt := ng.CreateVote(apirequest.CreateVote{
		Name:              fmt.Sprintf("%s_V", t.Name()),
		VoteOptions:       []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
		IdleTimeoutSecond: 300,
	}).Join()
	if ovt == nil {
		assert.Fail(t, "create vt nil")
		return
	}

	println(ovt.Id())
	thread := 100
	times := 20
	bwg.Add(thread)
	wjd := concurrent.WaitGroup{}
	wjd.Add(thread)
	wcd := concurrent.WaitGroup{}
	wcd.Add(thread)

	go func() {
		<-time.After(time.Second * 10)
		if c := wjd.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("wjd %d", c))
		}

		wjd.Reset()
	}()

	for i := 0; i < thread; i++ {
		go func(ii int) {
			og := NewOmegaBuilder(conf).Connect().Get()
			ch := og.Channel(och.Id()).Join("")
			if ch == nil {
				assert.Fail(t, "get ch nil")
				return
			}

			vt := og.Vote(ovt.Id()).Join("")
			if vt == nil {
				assert.Fail(t, "get vt nil")
				return
			}

			wjd.Done()
			wjd.Wait()
			on := int32(0)
			wrd := concurrent.WaitGroup{}
			wrd.Add(thread * times)

			go func(i int) {
				<-time.After(time.Second * 30)
				if c := wrd.Remain(); c > 0 {
					assert.Fail(t, fmt.Sprintf("wrd %d", c))
					println(fmt.Sprintf("%d timeout %d", ii, on))
				}
			}(ii)

			ch.Watch(func(msg messages.TransitFrame) {
				if _, ok := msg.(*messages.ChannelMessage); ok {
					atomic.AddInt32(&on, 1)
					atomic.AddInt32(&nCount, 1)
					wrd.Done()
				}
			})

			for ir := 0; ir < times; ir++ {
				time.Sleep(time.Millisecond * 100)
				if !ch.SendMessage(fmt.Sprintf("%d !!!", ir), nil).AwaitTimeout(5 * Timeout).IsSuccess() {
					assert.Fail(t, "send fail")
				}

				if !vt.VoteOptions()[rand.Int()%2].Select() {
					assert.Fail(t, "select fail")
				}

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
		<-time.After(time.Second * 30)
		if c := bwg.Remain(); c > 0 {
			assert.Fail(t, fmt.Sprintf("bwg %d burst", c))
		}

		bwg.Reset()
	}()

	bwg.Wait()
	assert.Equal(t, int32(thread*thread*times), nCount)
	println(nCount)
	<-time.After(time.Second)
	vtc := ovt.Count().TransitFrame()
	assert.Equal(t, 1, int(vtc.GetVoteOptions()[0].Count+vtc.GetVoteOptions()[1].Count))
	och.Close()
	assert.True(t, ovt.Close().AwaitTimeout(Timeout).IsSuccess())
	ng.Close().Await()
}
