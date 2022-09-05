package base

import (
	"fmt"
	"os"
	"testing"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/base/apiresponse"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	appId := os.Getenv("TEST_APP_ID")
	token := os.Getenv("TEST_TOKEN")
	domain := os.Getenv("TEST_DOMAIN")
	conf := NewConfig(appId, token)
	conf.Domain = domain
	e := NewEngine(conf)
	rmsg := "test message"
	remoteSession := e.connect().Session()
	session := e.connect().Session()
	assert.NotEmpty(t, session)
	assert.False(t, session.isClosed())
	assert.NotEmpty(t, session.GetId())
	assert.NotEmpty(t, session.ping().Get())
	rAgent := NewAgent(remoteSession)
	agent := NewAgent(session)
	rs := agent.GetRemoteSession(remoteSession.GetId()).Session()
	assert.Equal(t, remoteSession.GetId(), rs.GetId())
	remoteSession.OnMessage(func(msg *omega.TransitFrame) {
		println(fmt.Sprintf("remote get to remote: %s", msg.GetSessionMessage().Message))
		assert.Equal(t, rmsg, msg.GetSessionMessage().Message)
	})

	rs.SendMessage(rmsg).Await()
	agent.SessionMessage(remoteSession.GetId(), rmsg).Await()

	rs.OnMessage(func(msg *omega.TransitFrame) {
		println(fmt.Sprintf("local get to local: %s", msg.GetSessionMessage().Message))
		assert.Equal(t, rmsg, msg.GetSessionMessage().Message)
	})

	rAgent.GetRemoteSession(session.GetId()).Session().SendMessage(rmsg)
	rAgent.SessionMessage(session.GetId(), rmsg).Await()
	if f := e.createChannel(apirequest.CreateChannel{Name: "kumoi-agent-golang_test_ch"}).Await(); f.Get() != nil {
		resp := f.Get().(*apiresponse.CreateChannel)
		assert.Equal(t, "kumoi-agent-golang_test_ch", resp.Name)
		assert.NotEmpty(t, resp.AppId)
		assert.NotEmpty(t, resp.ChannelId)
		assert.NotEmpty(t, resp.OwnerKey)
		assert.NotEmpty(t, resp.ParticipatorKey)
		assert.Equal(t, 60, resp.IdleTimeoutSecond)
		assert.Empty(t, agent.CloseChannel(resp.ChannelId, "").Get())
		assert.NotEmpty(t, agent.CloseChannel(resp.ChannelId, resp.OwnerKey).Get())
		assert.Empty(t, agent.CloseChannel(resp.ChannelId, resp.OwnerKey).Get())
	} else {
		assert.Error(t, f.Error())
	}

	if f := e.createVote(apirequest.CreateVote{Name: "kumoi-agent-golang_test_vt", VoteOptions: []apirequest.CreateVoteOption{{Name: "vto1"}, {Name: "vto2"}}}).Await(); f.Get() != nil {
		resp := f.Get().(*apiresponse.CreateVote)
		assert.Equal(t, "kumoi-agent-golang_test_vt", resp.Name)
		assert.NotEmpty(t, resp.AppId)
		assert.NotEmpty(t, resp.VoteId)
		assert.NotEmpty(t, resp.Key)
		assert.Equal(t, 2, len(resp.VoteOptions))
		assert.Equal(t, "vto1", resp.VoteOptions[0].Name)
		assert.NotEmpty(t, resp.VoteOptions[0].Id)
		assert.Equal(t, 60, resp.IdleTimeoutSecond)
		assert.Empty(t, agent.CloseVote(resp.VoteId, "").Get())
		assert.NotEmpty(t, agent.CloseVote(resp.VoteId, resp.Key).Get())
		assert.Empty(t, agent.CloseVote(resp.VoteId, resp.Key).Get())
	} else {
		assert.Error(t, f.Error())
	}

	session.Disconnect().Await()
	assert.True(t, session.isClosed())

	bwg := concurrent.WaitGroup{}
	for i := 0; i < 10; i++ {
		bwg.Add(1)
		go func(i int) {
			session := e.connect().Session()
			assert.NotEmpty(t, session)
			assert.False(t, session.isClosed())
			session.Disconnect().Await()
			println(fmt.Sprintf("session connected %d", i))
			bwg.Done()
		}(i)
	}

	go func() {
		<-time.After(10 * time.Second)
		bwg.Reset()
	}()

	remoteSession.Disconnect().Await()
	bwg.Wait()
}
