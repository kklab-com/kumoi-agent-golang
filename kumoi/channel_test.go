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
	o := NewOmegaBuilder(engine).Connect().Omega()
	assert.NotEmpty(t, o)
	assert.NotEmpty(t, o.Hello())
	assert.NotEmpty(t, o.Time())
	chf := o.CreateChannel(apirequest.CreateChannel{
		Name:              t.Name(),
		IdleTimeoutSecond: 100,
	})

	assert.NotNil(t, chf.Response())
	chInfo := chf.Info()
	assert.NotNil(t, chInfo)
	assert.Equal(t, t.Name(), chInfo.Name())
	ch := chInfo.Join(chf.Response().ParticipatorKey)
	leave := false
	ch.OnLeave(func() {
		leave = true
	})

	cmMeta := concurrent.NewFuture()
	ch.Watch(func(msg messages.ChannelFrame) {
		println(reflect.ValueOf(msg).Elem().Type().Name())
		switch v := msg.(type) {
		case *messages.ChannelMessage:
			if v.Metadata != nil && v.Metadata.GetFields()["typ"].GetStringValue() == "SendMessage" {
				cmMeta.Completable().Complete(true)
			}
		}
	})

	assert.NotNil(t, ch)
	assert.NotEqual(t, ch.Role(), omega.Role_RoleOwner)
	assert.True(t, ch.SendMessage("SendMessage", base.NewMetadata(map[string]interface{}{"typ": "SendMessage"})))
	assert.False(t, ch.SendOwnerMessage("SendOwnerMessage", nil))
	assert.True(t, cmMeta.GetTimeout(time.Second).(bool))
	assert.False(t, ch.Close())
	assert.Nil(t, chInfo.Join(""))
	assert.True(t, ch.Leave())
	assert.True(t, leave)

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
	assert.True(t, ch.SendOwnerMessage("SendOwnerMessage", nil))
	assert.True(t, ch.SetName("new_channel_name"))
	println("wait for replay")
	<-time.After(2 * time.Second)
	cp := ch.ReplayChannelMessage(0, true, omega.Volume_VolumeLowest)
	assert.NotNil(t, cp)
	assert.Equal(t, int32(1), ch.GetCount().Count)
	assert.Equal(t, ch.Id(), cp.Next().(*ChannelPlayerEntity).GetGetChannelMeta().GetChannelId())
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
	assert.Equal(t, ch.Id(), cp.Next().(*ChannelPlayerEntity).GetGetChannelMeta().GetChannelId())
	assert.Equal(t, ch.Id(), cp.Next().(*ChannelPlayerEntity).GetCloseChannel().GetChannelId())
	assert.Nil(t, cp.Next())
	assert.True(t, closed)
	assert.True(t, o.Close().Await().IsSuccess())
}
