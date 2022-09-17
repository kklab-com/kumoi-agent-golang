package kumoi

import (
	"testing"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Broadcast(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()
	oSecond.OnBroadcastHandler = func(tf *messages.Broadcast) {
		if tf.GetMessage() == TestMessage {
			okFuture.Completable().Complete(nil)
		}
	}

	oFirst.Broadcast(TestMessage)
	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_ChannelCount(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Response()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.OwnerKey)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	ch := reqFuture.Join()
	assert.NotEmpty(t, ch)

	ch2 := oSecond.Channel(resp.ChannelId).Join("")
	assert.NotEmpty(t, ch2)
	assert.True(t, ch.Count().TransitFrame().GetCount() > 0)

	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())

	okFuture.Completable().Complete(nil)
	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_ChannelMessage(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Response()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.OwnerKey)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	ch := reqFuture.Join()
	assert.NotEmpty(t, ch)

	ch2 := oSecond.Channel(resp.ChannelId).Join("")
	assert.NotEmpty(t, ch2)
	ch2.Watch(func(msg messages.TransitFrame) {
		switch v := msg.(type) {
		case *messages.ChannelMessage:
			println(v.ProtoMessage().String())
			if v.GetMessage() == TestMessage &&
				v.GetFromSession() == oFirst.Session().GetId() &&
				v.GetMetadata().AsMap()["MK"] == "MV" {
				okFuture.Completable().Complete(nil)
			}
		}
	})

	assert.True(t, ch.SendMessage(TestMessage, base.NewMetadata(map[string]interface{}{"MK": "MV"})).AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_ChannelOwnerMessage(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Response()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.OwnerKey)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	ch := reqFuture.Join()
	assert.NotEmpty(t, ch)

	ch2 := oSecond.Channel(resp.ChannelId).Join("")
	assert.NotEmpty(t, ch2)
	ch2.Watch(func(msg messages.TransitFrame) {
		switch v := msg.(type) {
		case *messages.ChannelOwnerMessage:
			if v.GetMessage() == TestMessage &&
				v.GetFromSession() == oFirst.Session().GetId() &&
				v.GetMetadata().AsMap()["MK"] == "MV" {
				okFuture.Completable().Complete(nil)
			}
		}
	})

	assert.True(t, ch.SendOwnerMessage(TestMessage, base.NewMetadata(map[string]interface{}{"MK": "MV"})).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, ch2.SendOwnerMessage(TestMessage, nil).AwaitTimeout(Timeout).IsFail())

	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_GetChannelMeta(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Response()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.OwnerKey)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	ch := reqFuture.Join()
	assert.NotEmpty(t, ch)

	assert.True(t, ch.SetMetadata(base.NewMetadata(map[string]interface{}{"MK": "MV"})).AwaitTimeout(Timeout).IsSuccess())
	ch2 := oSecond.Channel(resp.ChannelId).Join("")
	assert.NotEmpty(t, ch2)
	assert.True(t, ch2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetData().AsMap()["MK"] == "MV")
	assert.True(t, ch2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetName() == t.Name())
	okFuture.Completable().Complete(nil)

	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_GetSessionMeta(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	assert.True(t, oFirst.Session().SetMetadata(base.NewMetadata(map[string]interface{}{"MK": "MV"})).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Session().Fetch().AwaitTimeout(Timeout).TransitFrame().GetData().AsMap()["MK"] == "MV")
	assert.True(t, oSecond.GetRemoteSession(oFirst.Session().GetId()).GetMetadata().AsMap()["MK"] == "MV")
	okFuture.Completable().Complete(nil)

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_GetVoteMeta(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateVote(apirequest.CreateVote{
		Name:        t.Name(),
		VoteOptions: []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
	})

	resp := reqFuture.Response()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.Key)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	vt := reqFuture.Join()
	assert.NotEmpty(t, vt)

	assert.True(t, vt.SetMetadata(base.NewMetadata(map[string]interface{}{"MK": "MV"})).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Info().VoteOptions()[0].Name == "vto1")
	assert.True(t, vt.Info().VoteOptions()[1].Name == "vto2")
	assert.True(t, vt.Info().VoteOptions()[0].Id != "")
	assert.True(t, vt.Info().VoteOptions()[1].Id != "")
	vt2 := oSecond.Vote(resp.VoteId).Join("")
	assert.NotEmpty(t, vt2)
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetData().AsMap()["MK"] == "MV")
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetName() == t.Name())
	okFuture.Completable().Complete(nil)

	assert.True(t, vt.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_Hello(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	assert.True(t, oFirst.Hello().AwaitTimeout(Timeout).TransitFrame().GetSessionId() == oFirst.Session().GetId())
	assert.True(t, oFirst.Hello().AwaitTimeout(Timeout).TransitFrame().GetSubject() == oFirst.Session().GetSubject())
	assert.True(t, oFirst.Hello().AwaitTimeout(Timeout).TransitFrame().GetSubjectName() == oFirst.Session().GetName())
	assert.True(t, oSecond.Hello().AwaitTimeout(Timeout).TransitFrame().GetSessionId() == oSecond.Session().GetId())
	assert.True(t, oSecond.Hello().AwaitTimeout(Timeout).TransitFrame().GetSubject() == oSecond.Session().GetSubject())
	assert.True(t, oSecond.Hello().AwaitTimeout(Timeout).TransitFrame().GetSubjectName() == oSecond.Session().GetName())
	okFuture.Completable().Complete(nil)

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_ChannelLeaveChannel(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Response()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.OwnerKey)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	ch := reqFuture.Join()
	assert.NotEmpty(t, ch)
	assert.Nil(t, reqFuture.Join())
	ch2 := oSecond.Channel(resp.ChannelId).Join("")
	assert.NotEmpty(t, ch2)
	assert.Nil(t, oSecond.Channel(resp.ChannelId).Join(""))
	assert.True(t, ch.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, ch.Leave().AwaitTimeout(Timeout).IsFail())
	assert.True(t, ch2.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, ch2.Leave().AwaitTimeout(Timeout).IsFail())

	ch = oFirst.Channel(resp.ChannelId).Join("")
	assert.NotEmpty(t, ch)
	assert.True(t, ch.Leave().AwaitTimeout(Timeout).IsSuccess())
	ch = reqFuture.Join()

	okFuture.Completable().Complete(nil)

	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_LeaveVote(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateVote(apirequest.CreateVote{
		Name:        t.Name(),
		VoteOptions: []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
	})

	resp := reqFuture.Response()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.Key)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	vt := reqFuture.Join()
	assert.NotEmpty(t, vt)

	assert.True(t, vt.SetMetadata(base.NewMetadata(map[string]interface{}{"MK": "MV"})).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Info().VoteOptions()[0].Name == "vto1")
	assert.True(t, vt.Info().VoteOptions()[1].Name == "vto2")
	assert.True(t, vt.Info().VoteOptions()[0].Id != "")
	assert.True(t, vt.Info().VoteOptions()[1].Id != "")
	vt2 := oSecond.Vote(resp.VoteId).Join("")
	assert.NotEmpty(t, vt2)
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetData().AsMap()["MK"] == "MV")
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetName() == t.Name())
	okFuture.Completable().Complete(nil)

	assert.True(t, vt2.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_ServerTime(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	assert.True(t, oFirst.ServerTime().AwaitTimeout(Timeout).TransitFrame().GetUnixNano() != 0)
	assert.True(t, oFirst.ServerTime().AwaitTimeout(Timeout).TransitFrame().GetUnixTimestamp() != 0)
	assert.True(t, oSecond.ServerTime().AwaitTimeout(Timeout).TransitFrame().GetUnixNano() != 0)
	assert.True(t, oSecond.ServerTime().AwaitTimeout(Timeout).TransitFrame().GetUnixTimestamp() != 0)
	okFuture.Completable().Complete(nil)

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_SessionMessage(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Omega()
	oSecond := NewOmegaBuilder(conf).Connect().Omega()
	okFuture := concurrent.NewFuture()

	oSecond.GetRemoteSession(oFirst.Session().GetId()).OnMessage(func(msg *messages.SessionMessage) {
		if msg.GetMessage() == TestMessage {
			okFuture.Completable().Complete(nil)
		}
	})

	assert.True(t, oFirst.GetRemoteSession(oSecond.Session().GetId()).SendMessage(TestMessage).AwaitTimeout(Timeout).TransitFrame().GetTransitId() > 0)

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}
