package kumoi

import (
	"testing"
	"time"

	"github.com/kklab-com/goth-kkutil/concurrent"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	"github.com/stretchr/testify/assert"
)

func TestOmega_RemoteSession(t *testing.T) {
	ro := NewOmegaBuilder(conf).Connect().Omega()
	so := NewOmegaBuilder(conf).Connect().Omega()
	assert.NotNil(t, ro)
	assert.NotNil(t, so)
	got := false
	gc := 0

	bwg := concurrent.BurstWaitGroup{}
	ro.GetAgentSession().OnMessage(func(msg *messages.SessionMessage) {
		println(msg.Message)
		bwg.Done()
		got = true
		gc++
	})

	srs := so.GetRemoteSession(ro.GetAgentSession().GetId())
	assert.NotNil(t, srs)
	for i := 0; i < 10; i++ {
		bwg.Add(1)
		srs.SendMessage("send")
	}

	go func() {
		<-time.After(time.Second)
		bwg.Burst()
	}()

	bwg.Wait()
	assert.True(t, got)
	assert.Equal(t, 10, gc)
	ro.Close().Await()
	so.Close().Await()
}
