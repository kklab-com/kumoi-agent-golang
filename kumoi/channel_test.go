package kumoi

import (
	"math"
	"reflect"
	"testing"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

func TestChannel_ChannelJoin(t *testing.T) {
	omg := NewOmegaBuilder(conf).Connect().Omega()
	chf := omg.CreateChannel(apirequest.CreateChannel{
		Name:              t.Name(),
		IdleTimeoutSecond: 100,
	})

	assert.NotNil(t, chf.Response())
	chInfo := chf.Info()
	assert.NotNil(t, chInfo)
	assert.Equal(t, t.Name(), chInfo.Name())
	ch := chInfo.Join(chf.Response().ParticipatorKey)
	assert.NotNil(t, ch)
	vf := concurrent.NewFuture()
	ch.OnLeave(func() {
		vf.Completable().Complete(nil)
	})

	cmMeta := concurrent.NewFuture()
	ch.Watch(func(msg messages.ChannelFrame) {
		println(reflect.ValueOf(msg).Elem().Type().Name())
		switch v := msg.(type) {
		case *messages.ChannelMessage:
			if v.Metadata != nil && v.Metadata.GetFields()["typ"].GetStringValue() == "SendMessage" {
				cmMeta.Completable().Complete(nil)
			}
		}
	})

	assert.NotNil(t, ch)
	assert.NotEqual(t, ch.Role(), omega.Role_RoleOwner)
	assert.True(t, ch.SendMessage("SendMessage", base.NewMetadata(map[string]interface{}{"typ": "SendMessage"})).AwaitTimeout(Timeout).IsSuccess())
	assert.False(t, ch.SendOwnerMessage("SendOwnerMessage", nil).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, cmMeta.Await().IsSuccess())
	assert.False(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.Nil(t, chInfo.Join(""))
	assert.True(t, ch.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vf.AwaitTimeout(Timeout).IsSuccess())

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
	assert.True(t, ch.SendOwnerMessage("SendOwnerMessage", nil).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, ch.SetName("new_channel_name").AwaitTimeout(Timeout).IsSuccess())
	println("wait for replay")
	<-time.After(2 * time.Second)
	cp := ch.ReplayChannelMessage(0, true, omega.Volume_VolumeLowest)
	assert.NotNil(t, cp)
	assert.Equal(t, int32(1), ch.Count().TransitFrame().Count)
	assert.Equal(t, ch.Id(), cp.Next().Cast().GetChannelMeta().GetChannelId())
	assert.Equal(t, ch.Name(), cp.Next().Cast().SetChannelMeta().Name)
	assert.Equal(t, string((&omega.ChannelOwnerMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, string((&omega.ChannelMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Nil(t, cp.Next())
	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())
	println("wait for playback")
	<-time.After(2 * time.Second)
	cp = omg.PlaybackChannelMessage(ch.Id(), math.MaxInt32, false, omega.Volume_VolumeLowest)
	assert.NotNil(t, cp)
	assert.Equal(t, string((&omega.ChannelMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, string((&omega.ChannelOwnerMessage{}).ProtoReflect().Descriptor().Name()), cp.Next().TypeName())
	assert.Equal(t, ch.Name(), cp.Next().Cast().SetChannelMeta().Name)
	assert.Equal(t, ch.Id(), cp.Next().Cast().GetChannelMeta().GetChannelId())
	assert.Equal(t, ch.Id(), cp.Next().Cast().CloseChannel().GetChannelId())
	assert.Nil(t, cp.Next())
	assert.True(t, closed)
	assert.True(t, omg.Close().Await().AwaitTimeout(Timeout).IsSuccess())
}
