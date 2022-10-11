package kumoi

import (
	"testing"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Broadcast(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
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
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Get()
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
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Get()
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
			if v.GetMessage() == TestMessage &&
				v.GetFromSession() == oFirst.Session().GetId() &&
				v.GetMetadata()["MK"] == "MV" {
				okFuture.Completable().Complete(nil)
			}
		}
	})

	assert.True(t, ch.SendMessage(TestMessage, map[string]interface{}{"MK": "MV"}).AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_ChannelOwnerMessage(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Get()
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
				v.GetMetadata()["MK"] == "MV" {
				okFuture.Completable().Complete(nil)
			}
		}
	})

	assert.True(t, ch.SendOwnerMessage(TestMessage, map[string]interface{}{"MK": "MV"}).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, ch2.SendOwnerMessage(TestMessage, nil).AwaitTimeout(Timeout).IsFail())

	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_GetChannelMeta(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Get()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.OwnerKey)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	ch := reqFuture.Join()
	assert.NotEmpty(t, ch)

	assert.True(t, ch.SetMetadata(map[string]interface{}{"MK": "MV"}).AwaitTimeout(Timeout).IsSuccess())
	ch2 := oSecond.Channel(resp.ChannelId).Join("")
	assert.NotEmpty(t, ch2)
	assert.True(t, ch2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetData()["MK"] == "MV")
	assert.True(t, ch2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetName() == t.Name())
	okFuture.Completable().Complete(nil)

	assert.True(t, ch.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_GetSessionMeta(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	assert.True(t, oFirst.Session().SetMetadata(map[string]interface{}{"MK": "MV"}).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Session().Fetch().AwaitTimeout(Timeout).TransitFrame().GetData()["MK"] == "MV")
	assert.True(t, oSecond.GetRemoteSession(oFirst.Session().GetId()).GetMetadata()["MK"] == "MV")
	okFuture.Completable().Complete(nil)

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_GetVoteMeta(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateVote(apirequest.CreateVote{
		Name:        t.Name(),
		VoteOptions: []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
	})

	resp := reqFuture.Get()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.Key)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	vt := reqFuture.Join()
	assert.NotEmpty(t, vt)

	assert.True(t, vt.SetMetadata(map[string]interface{}{"MK": "MV"}).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Info().VoteOptions()[0].Name == "vto1")
	assert.True(t, vt.Info().VoteOptions()[1].Name == "vto2")
	assert.True(t, vt.Info().VoteOptions()[0].Id != "")
	assert.True(t, vt.Info().VoteOptions()[1].Id != "")
	vt2 := oSecond.Vote(resp.VoteId).Join("")
	assert.NotEmpty(t, vt2)
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetData()["MK"] == "MV")
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetName() == t.Name())
	okFuture.Completable().Complete(nil)

	assert.True(t, vt.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_Hello(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
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
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateChannel(apirequest.CreateChannel{Name: t.Name()})
	resp := reqFuture.Get()
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
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateVote(apirequest.CreateVote{
		Name:        t.Name(),
		VoteOptions: []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
	})

	resp := reqFuture.Get()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.Key)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	vt := reqFuture.Join()
	assert.NotEmpty(t, vt)

	assert.True(t, vt.SetMetadata(map[string]interface{}{"MK": "MV"}).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Info().VoteOptions()[0].Name == "vto1")
	assert.True(t, vt.Info().VoteOptions()[1].Name == "vto2")
	assert.True(t, vt.Info().VoteOptions()[0].Id != "")
	assert.True(t, vt.Info().VoteOptions()[1].Id != "")
	vt2 := oSecond.Vote(resp.VoteId).Join("")
	assert.NotEmpty(t, vt2)
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetData()["MK"] == "MV")
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetName() == t.Name())
	okFuture.Completable().Complete(nil)

	assert.True(t, vt2.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_ServerTime(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
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
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
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

func TestMessage_SessionMessageSelf(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	oFirst.Session().OnMessage(func(msg *messages.SessionMessage) {
		if msg.GetMessage() == TestMessage {
			okFuture.Completable().Complete(nil)
		}
	})

	assert.True(t, oFirst.Session().SendMessage(TestMessage).AwaitTimeout(Timeout).TransitFrame().GetTransitId() > 0)

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_VoteCount(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateVote(apirequest.CreateVote{
		Name:        t.Name(),
		VoteOptions: []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
	})

	resp := reqFuture.Get()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.Key)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	vt := reqFuture.Join()
	assert.NotEmpty(t, vt)

	assert.True(t, vt.SetMetadata(map[string]interface{}{"MK": "MV"}).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Info().VoteOptions()[0].Name == "vto1")
	assert.True(t, vt.Info().VoteOptions()[1].Name == "vto2")
	assert.True(t, vt.Info().VoteOptions()[0].Id != "")
	assert.True(t, vt.Info().VoteOptions()[1].Id != "")
	vt2 := oSecond.Vote(resp.VoteId).Join("")
	assert.NotEmpty(t, vt2)
	assert.True(t, vt2.Info().VoteOptions()[1].Select())
	assert.True(t, vt.Count().AwaitTimeout(Timeout).TransitFrame().GetVoteOptions()[0].GetCount() == 0)
	assert.True(t, vt.Count().AwaitTimeout(Timeout).TransitFrame().GetVoteOptions()[1].GetCount() > 0)

	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetData()["MK"] == "MV")
	assert.True(t, vt2.Fetch().AwaitTimeout(Timeout).TransitFrame().GetName() == t.Name())
	okFuture.Completable().Complete(nil)

	assert.True(t, vt2.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_VoteMessage(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateVote(apirequest.CreateVote{
		Name:        t.Name(),
		VoteOptions: []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
	})

	resp := reqFuture.Get()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.Key)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	vt := reqFuture.Join()
	assert.NotEmpty(t, vt)

	vt2 := oSecond.Vote(resp.VoteId).Join("")
	vt2.Watch(func(msg messages.TransitFrame) {
		switch v := msg.(type) {
		case *messages.VoteMessage:
			if v.GetMessage() == TestMessage {
				okFuture.Completable().Complete(nil)
			}
		}
	})

	assert.True(t, vt.SendMessage(TestMessage).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt2.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_VoteOwnerMessage(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateVote(apirequest.CreateVote{
		Name:        t.Name(),
		VoteOptions: []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
	})

	resp := reqFuture.Get()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.Key)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	vt := reqFuture.Join()
	assert.NotEmpty(t, vt)

	vt2 := oSecond.Vote(resp.VoteId).Join("")
	vt2.Watch(func(msg messages.TransitFrame) {
		switch v := msg.(type) {
		case *messages.VoteOwnerMessage:
			if v.GetMessage() == TestMessage {
				okFuture.Completable().Complete(nil)
			}
		}
	})

	assert.True(t, vt.SendOwnerMessage(TestMessage).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt2.SendOwnerMessage(TestMessage).AwaitTimeout(Timeout).IsFail())
	assert.True(t, vt2.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}

func TestMessage_VoteStatus(t *testing.T) {
	oFirst := NewOmegaBuilder(conf).Connect().Get()
	oSecond := NewOmegaBuilder(conf).Connect().Get()
	okFuture := concurrent.NewFuture()

	reqFuture := oFirst.CreateVote(apirequest.CreateVote{
		Name:        t.Name(),
		VoteOptions: []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
	})

	resp := reqFuture.Get()
	assert.NotEmpty(t, resp)
	assert.NotEmpty(t, resp.Key)
	info := reqFuture.Info()
	assert.NotEmpty(t, info)
	assert.Equal(t, t.Name(), info.Name())

	vt := reqFuture.Join()
	assert.NotEmpty(t, vt)

	vt2 := oSecond.Vote(resp.VoteId).Join("")
	vt2.Watch(func(msg messages.TransitFrame) {
		switch v := msg.(type) {
		case *messages.VoteStatus:
			if v.GetStatus() == omega.Vote_StatusDeny {
				okFuture.Completable().Complete(nil)
			}
		}
	})

	assert.True(t, vt2.Select(vt2.Info().VoteOptions()[0].Id))
	assert.True(t, !vt2.Select(""))
	assert.True(t, vt.Status(omega.Vote_StatusDeny).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Status(omega.Vote_StatusDeny).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Status(omega.Vote_StatusDeny).AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt2.Status(omega.Vote_StatusDeny).AwaitTimeout(Timeout).IsFail())
	assert.True(t, !vt2.Select(vt2.Info().VoteOptions()[0].Id))

	assert.True(t, vt2.Leave().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, vt.Close().AwaitTimeout(Timeout).IsSuccess())

	assert.True(t, okFuture.AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oFirst.Close().AwaitTimeout(Timeout).IsSuccess())
	assert.True(t, oSecond.Close().AwaitTimeout(Timeout).IsSuccess())
}
