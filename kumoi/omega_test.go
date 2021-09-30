package kumoi

import (
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/kklab-com/goth-kkutil/concurrent"
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

func TestMain(m *testing.M) {
	appId = os.Getenv("TEST_APP_ID")
	token = os.Getenv("TEST_TOKEN")
	domain = os.Getenv("TEST_DOMAIN")
	conf = base.NewConfig(appId, token)
	conf.Domain = domain
	m.Run()
}

func TestOmega(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Omega()
	assert.NotEmpty(t, o)
	assert.NotEmpty(t, o.Hello())
	assert.NotEmpty(t, o.Time())
	assert.True(t, o.Broadcast("golang broadcast test"))
	chf := o.CreateChannel(apirequest.CreateChannel{
		Name:              "!!!",
		IdleTimeoutSecond: 300,
	})

	assert.NotNil(t, chf.Response())
	chInfo := chf.Info()
	assert.NotNil(t, chInfo)
	ch := chInfo.Join(chf.Response().ParticipatorKey)
	leave := false
	ch.OnLeave(func() {
		leave = true
	})

	ch.Watch(func(msg messages.ChannelFrame) {
		println(reflect.ValueOf(msg).Elem().Type().Name())
	})

	assert.NotNil(t, ch)
	assert.NotEqual(t, ch.Role(), omega.Role_RoleOwner)
	assert.True(t, ch.SendMessage("SendMessage"))
	assert.False(t, ch.SendOwnerMessage("SendOwnerMessage"))
	assert.False(t, ch.Close())
	assert.Nil(t, chInfo.Join(""))
	assert.True(t, ch.Leave())

	ch = chf.Join()
	closed := false
	ch.OnClose(func() {
		closed = true
	})

	ch.Watch(func(msg messages.ChannelFrame) {
		println(reflect.ValueOf(msg).Elem().Type().Name())
	})

	assert.NotNil(t, ch)
	assert.Equal(t, ch.Role(), omega.Role_RoleOwner)
	assert.True(t, ch.SendOwnerMessage("SendOwnerMessage"))
	assert.True(t, ch.SetName("new_channel_name"))
	println("wait for replay")
	<-time.After(2 * time.Second)
	cp := ch.ReplayChannelMessage(0, true, omega.Volume_VolumeLowest)
	assert.NotNil(t, cp)
	assert.Equal(t, int32(1), ch.GetCount().Count)
	assert.Equal(t, ch.Id(), cp.Next().(*ChannelPlayerEntity).GetGetChannelMeta().ChannelId())
	assert.Equal(t, ch.Name(), cp.Next().(*ChannelPlayerEntity).GetSetChannelMeta().Name)
	assert.Equal(t, string((&omega.ChannelOwnerMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().Name())
	assert.Equal(t, string((&omega.ChannelMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().Name())
	assert.Nil(t, cp.Next())
	assert.True(t, ch.Close())
	println("wait for playback")
	<-time.After(2 * time.Second)
	cp = o.PlaybackChannelMessage(ch.Id(), math.MaxInt32, false, omega.Volume_VolumeLowest)
	assert.NotNil(t, cp)
	assert.Equal(t, string((&omega.ChannelMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().Name())
	assert.Equal(t, string((&omega.ChannelOwnerMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().Name())
	assert.Equal(t, ch.Name(), cp.Next().(*ChannelPlayerEntity).GetSetChannelMeta().Name)
	assert.Equal(t, ch.Id(), cp.Next().(*ChannelPlayerEntity).GetGetChannelMeta().ChannelId())
	assert.Equal(t, ch.Id(), cp.Next().(*ChannelPlayerEntity).GetCloseChannel().ChannelId())
	assert.Nil(t, cp.Next())
	assert.True(t, leave)
	assert.True(t, closed)
	assert.True(t, o.Close().Await().IsSuccess())
}

func TestOmegaDisconnect(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Omega()
	bwg := concurrent.BurstWaitGroup{}
	go func() {
		<-time.After(3 * time.Second)
		if bwg.Remain() > 0 {
			bwg.Burst()
			assert.Fail(t, "disconnected not invoke")
		}
	}()

	bwg.Add(1)
	o.OnDisconnectedHandler = func() {
		bwg.Done()
	}

	go func() {
		assert.False(t, o.IsClosed())
		assert.False(t, o.IsDisconnected())
		<-time.After(1 * time.Second)
		o.Close().Await()
		assert.True(t, o.IsClosed())
	}()

	bwg.Wait()
	assert.True(t, o.IsClosed())
	assert.True(t, o.IsDisconnected())
}

func TestOmegaWriteOnClosed(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Omega()
	chResp := o.CreateChannel(apirequest.CreateChannel{})
	chInfo := chResp.Info()
	ch := chInfo.Join("")
	assert.True(t, ch.SendMessage("!!!"))
	o.Close().Await()
	assert.False(t, ch.SendMessage("!!!"))
}
