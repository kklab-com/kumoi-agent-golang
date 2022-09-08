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
var engine *base.Engine

const Timeout = time.Second

func TestMain(m *testing.M) {
	appId = os.Getenv("TEST_APP_ID")
	token = os.Getenv("TEST_TOKEN")
	domain = os.Getenv("TEST_DOMAIN")
	conf := base.NewConfig(appId, token)
	conf.Domain = domain
	engine = base.NewEngine(conf)
	m.Run()
}

func TestOmega(t *testing.T) {
	o := NewOmegaBuilder(engine).Connect().Omega()
	assert.NotEmpty(t, o)
	assert.NotEmpty(t, o.Hello())
	assert.NotEmpty(t, o.ServerTime())
	assert.True(t, o.Broadcast("golang broadcast test"))
	chf := o.CreateChannel(apirequest.CreateChannel{
		Name:              "!!!",
		IdleTimeoutSecond: 300,
	})

	assert.NotNil(t, chf.Response())
	chInfo := chf.Info()
	assert.NotNil(t, chInfo)
	ch := chInfo.Join(chf.Response().ParticipatorKey)
	vfl := concurrent.NewFuture()
	ch.OnLeave(func() {
		vfl.Completable().Complete(nil)
	})

	ch.Watch(func(msg messages.ChannelFrame) {
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

	ch.Watch(func(msg messages.ChannelFrame) {
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
	assert.Equal(t, int32(1), ch.GetCount().TransitFrame().Count)
	assert.Equal(t, ch.Id(), cp.Next().Cast().GetChannelMeta().GetChannelId())
	assert.Equal(t, ch.Name(), cp.Next().Cast().SetChannelMeta().Name)
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
	assert.Equal(t, ch.Name(), cp.Next().Cast().SetChannelMeta().Name)
	assert.Equal(t, ch.Id(), cp.Next().Cast().GetChannelMeta().GetChannelId())
	assert.Equal(t, ch.Id(), cp.Next().Cast().CloseChannel().GetChannelId())
	assert.Nil(t, cp.Next())
	assert.True(t, vfc.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, o.Close().Await().AwaitTimeout(Timeout).IsSuccess())
}

func TestOmegaClose(t *testing.T) {
	o := NewOmegaBuilder(engine).Connect().Omega()
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
	o := NewOmegaBuilder(engine).Connect().Omega()
	chResp := o.CreateChannel(apirequest.CreateChannel{})
	chInfo := chResp.Info()
	ch := chInfo.Join("")
	assert.True(t, ch.SendMessage("!!!", nil).AwaitTimeout(Timeout).IsSuccess())
	o.Close().Await()
	assert.False(t, ch.SendMessage("!!!", nil).AwaitTimeout(Timeout).IsSuccess())
}

func TestOmega_MultiVoteChannel(t *testing.T) {
	tCount := int32(0)
	nCount := int32(0)
	bwg := concurrent.WaitGroup{}
	ng := NewOmegaBuilder(engine).Connect().Omega()
	och := ng.CreateChannel(apirequest.CreateChannel{Name: "TestOmega_MultiVoteChannel_C"}).Join()
	if och == nil {
		assert.Fail(t, "create ch nil")
		return
	}

	ovt := ng.CreateVote(apirequest.CreateVote{
		Name:              "TestOmega_MultiVoteChannel_V",
		VoteOptions:       []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
		IdleTimeoutSecond: 300,
	}).Join()
	if ovt == nil {
		assert.Fail(t, "create vt nil")
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
			og := NewOmegaBuilder(engine).Connect().Omega()
			ch := og.GetChannel(och.Id()).Join("")
			if ch == nil {
				assert.Fail(t, "get ch nil")
				return
			}

			vt := og.GetVote(ovt.Id()).Join("")
			if vt == nil {
				assert.Fail(t, "get vt nil")
				return
			}

			wjd.Done()
			wjd.Wait()
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
				if !ch.SendMessage(fmt.Sprintf("%d !!!", ir), nil).AwaitTimeout(Timeout).IsSuccess() {
					assert.Fail(t, "send fail")
				}

				if !vt.Info().VoteOptions()[rand.Int()%2].Select() {
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
	assert.Equal(t, int32(thread*((thread+2)*times)), tCount)
	assert.Equal(t, int32(thread*thread*times), nCount)
	println(tCount)
	println(nCount)
	<-time.After(time.Second)
	vtc := ovt.GetCount().TransitFrame()
	assert.Equal(t, 1, int(vtc.VoteOptions[0].Count+vtc.VoteOptions[1].Count))
	och.Close()
	ovt.Close()
	ng.Close().Await()
}
