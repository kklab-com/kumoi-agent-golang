package kumoi

import (
	"testing"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	"github.com/stretchr/testify/assert"
)

func TestOmega_Session(t *testing.T) {
	o := NewOmegaBuilder(conf).Connect().Get()
	o.Close().Await()
}

func TestOmega_RemoteSession(t *testing.T) {
	ro := NewOmegaBuilder(conf).Connect().Get()
	so := NewOmegaBuilder(conf).Connect().Get()
	assert.NotNil(t, ro)
	assert.NotNil(t, so)
	got := false
	gc := 0

	bwg := concurrent.WaitGroup{}
	ro.Session().OnMessage(func(msg *messages.SessionMessage) {
		println(msg.GetMessage())
		bwg.Done()
		got = true
		gc++
	})

	srs := so.GetRemoteSession(ro.Session().GetId())
	assert.NotNil(t, srs)
	for i := 0; i < 10; i++ {
		bwg.Add(1)
		srs.SendMessage("send")
	}

	go func() {
		<-time.After(time.Second)
		bwg.Reset()
	}()

	bwg.Wait()
	assert.True(t, got)
	assert.Equal(t, 10, gc)
	ro.Close().Await()
	so.Close().Await()
}
