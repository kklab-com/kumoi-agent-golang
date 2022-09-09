package base

import (
	"testing"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

func TestAgentMessage(t *testing.T) {
	rmsg := "test message"
	rag := agentBuilder.Connect().Agent()
	ag := agentBuilder.Connect().Agent()
	assert.NotEmpty(t, ag)
	assert.False(t, ag.Session().IsClosed())
	assert.NotEmpty(t, ag.Session().GetId())
	assert.NotEmpty(t, ag.Ping().Get())
	vf := concurrent.NewFuture()
	rag.OnSessionMessage(func(msg *omega.TransitFrame) {
		assert.Equal(t, rmsg, msg.GetSessionMessage().Message)
		vf.Completable().Complete(nil)
	})

	ag.SessionMessage(rag.Session().GetId(), rmsg)
	assert.True(t, vf.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, ag.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, rag.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestAgent(t *testing.T) {
	rag := agentBuilder.Connect().Agent()
	ag := agentBuilder.Connect().Agent()
	assert.NotEmpty(t, ag)
	assert.False(t, ag.Session().IsClosed())
	assert.NotEmpty(t, ag.Session().GetId())
	assert.NotEmpty(t, ag.Ping().Get())

	if f := ag.CreateChannel(apirequest.CreateChannel{Name: "kumoi-agent-golang_test_ch", IdleTimeoutSecond: 60}).AwaitTimeout(Timeout); f.Get() != nil {
		resp := f.Get()
		assert.Equal(t, "kumoi-agent-golang_test_ch", resp.Name)
		assert.NotEmpty(t, resp.AppId)
		assert.NotEmpty(t, resp.ChannelId)
		assert.NotEmpty(t, resp.OwnerKey)
		assert.NotEmpty(t, resp.ParticipatorKey)
		assert.Equal(t, 60, resp.IdleTimeoutSecond)
		assert.Equal(t, omega.Role_RoleOwner, ag.JoinChannel(resp.ChannelId, resp.OwnerKey).TransitFrame().GetJoinChannel().GetRoleIndicator())
		assert.Equal(t, omega.Role_RoleParticipator, rag.JoinChannel(resp.ChannelId, "").TransitFrame().GetJoinChannel().GetRoleIndicator())
		vf1, vf2, vf3, vf4 := concurrent.NewFuture(), concurrent.NewFuture(), concurrent.NewFuture(), concurrent.NewFuture()
		rag.OnMessage(func(tf *omega.TransitFrame) {
			if msg := tf.GetChannelOwnerMessage(); msg != nil && msg.Message == "OWNER" {
				vf1.Completable().Complete(nil)
			}
		})

		rag.OnNotification(func(tf *omega.TransitFrame) {
			if msg := tf.GetChannelOwnerMessage(); msg != nil && msg.Message == "OWNER" {
				vf2.Completable().Complete(nil)
			}
		})

		ag.OnResponse(func(tf *omega.TransitFrame) {
			if tf.GetSetChannelMeta() != nil {
				vf3.Completable().Complete(nil)
			}
		})

		ag.OnNotification(func(tf *omega.TransitFrame) {
			if msg := tf.GetGetChannelMeta(); msg != nil && msg.GetName() == "NEW_CHANNEL" {
				vf4.Completable().Complete(nil)
			}
		})

		ag.ChannelOwnerMessage(resp.ChannelId, "OWNER", nil).AwaitTimeout(Timeout)
		assert.True(t, vf1.AwaitTimeout(Timeout).IsSuccess())
		assert.True(t, vf2.AwaitTimeout(Timeout).IsSuccess())
		assert.NotEmpty(t, ag.SetChannelMetadata(resp.ChannelId, "NEW_CHANNEL", nil, nil).GetTimeout(Timeout))
		assert.True(t, vf3.AwaitTimeout(Timeout).IsSuccess())
		assert.True(t, vf4.AwaitTimeout(Timeout).IsSuccess())
		assert.NotEmpty(t, rag.LeaveChannel(resp.ChannelId).GetTimeout(Timeout))
		assert.NotEmpty(t, ag.CloseChannel(resp.ChannelId, "").AwaitTimeout(Timeout).Error())
		assert.NotEmpty(t, ag.CloseChannel(resp.ChannelId, resp.OwnerKey).GetTimeout(Timeout))
		assert.NotEmpty(t, ag.CloseChannel(resp.ChannelId, resp.OwnerKey).AwaitTimeout(Timeout).Error())
	} else {
		assert.Error(t, f.Error())
	}

	if f := ag.CreateVote(apirequest.CreateVote{Name: "kumoi-agent-golang_test_vt", VoteOptions: []apirequest.CreateVoteOption{{Name: "vto1"}, {Name: "vto2"}}, IdleTimeoutSecond: 60}).AwaitTimeout(Timeout); f.Get() != nil {
		resp := f.Get()
		assert.Equal(t, "kumoi-agent-golang_test_vt", resp.Name)
		assert.NotEmpty(t, resp.AppId)
		assert.NotEmpty(t, resp.VoteId)
		assert.NotEmpty(t, resp.Key)
		assert.Equal(t, 2, len(resp.VoteOptions))
		assert.Equal(t, "vto1", resp.VoteOptions[0].Name)
		assert.NotEmpty(t, resp.VoteOptions[0].Id)
		assert.Equal(t, 60, resp.IdleTimeoutSecond)
		assert.NotEmpty(t, ag.CloseVote(resp.VoteId, "").AwaitTimeout(Timeout).Error())
		assert.NotEmpty(t, ag.CloseVote(resp.VoteId, resp.Key).GetTimeout(Timeout))
		assert.NotEmpty(t, ag.CloseVote(resp.VoteId, resp.Key).AwaitTimeout(Timeout).Error())
	} else {
		assert.Error(t, f.Error())
	}

	assert.True(t, ag.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, rag.Close().AwaitTimeout(Timeout).IsSuccess())
}
