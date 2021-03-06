package kumoi

import (
	"testing"
	"time"

	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

func TestOmegaVote(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Omega()
	assert.NotEmpty(t, o)
	assert.NotEmpty(t, o.Hello())
	assert.NotEmpty(t, o.Time())
	assert.True(t, o.Broadcast("golang broadcast test"))
	vtf := o.CreateVote(apirequest.CreateVote{
		Name:              "!!!",
		VoteOptions:       []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
		IdleTimeoutSecond: 300,
	})

	assert.NotNil(t, vtf.Get())
	vtInfo := vtf.Info()
	assert.NotNil(t, vtInfo)
	vt := vtInfo.Join("")
	vt.OnLeave(func() {
		println("first leave")
	})

	vt.Watch(func(msg messages.VoteFrame) {
		print("w")
	})

	assert.NotNil(t, vt)
	assert.True(t, vt.SendMessage("SendMessage"))
	assert.False(t, vt.SendOwnerMessage("SendOwnerMessage"))
	assert.False(t, vt.Status(omega.Vote_StatusDeny))
	assert.True(t, vt.Select(vt.Info().VoteOptions()[1].Id))
	assert.False(t, vtInfo.Close("").IsSuccess())
	assert.Nil(t, vtInfo.Join(""))
	assert.True(t, vt.Leave())

	vt = vtf.Join()
	vt.OnClose(func() {
		println("first close")
	})

	vt.Watch(func(msg messages.VoteFrame) {
		print("w")
	})

	assert.NotNil(t, vt)
	assert.True(t, vt.SendOwnerMessage("SendOwnerMessage"))
	assert.True(t, vt.SetName("new_vote_name"))
	assert.True(t, vt.Info().VoteOptions()[0].Select())
	assert.True(t, vt.Status(omega.Vote_StatusDeny))
	assert.False(t, vt.Select(vt.Info().VoteOptions()[1].Id))
	assert.Equal(t, int32(1), vt.GetCount().VoteOptions[0].Count)
	assert.True(t, vt.Close())
	assert.True(t, o.Close().Await().IsSuccess())
}

func TestOmegaVoteWatch(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Omega()
	assert.NotEmpty(t, o)
	vtf := o.CreateVote(apirequest.CreateVote{
		Name:              "!!!",
		VoteOptions:       []apirequest.CreateVoteOption{{"vto1"}, {"vto2"}},
		IdleTimeoutSecond: 300,
	})

	assert.NotNil(t, vtf)
	assert.NotNil(t, vtf.Response())
	vt := vtf.Join()
	vmCount := 0
	vt.Watch(func(msg messages.VoteFrame) {
		if _, ok := msg.(*messages.VoteMessage); ok {
			vmCount++
		}
	})

	assert.NotNil(t, vt)
	assert.True(t, vt.SendMessage("SendMessage"))
	time.Sleep(time.Second)
	assert.Equal(t, 1, vmCount)
	assert.True(t, vt.Close())
	assert.True(t, o.Close().Await().IsSuccess())
}
