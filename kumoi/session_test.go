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

func TestOmega_SelfRemoteSession(t *testing.T) {
	omg := NewOmegaBuilder(conf).Connect().Get()
	assert.NotNil(t, omg)
	rs := omg.GetRemoteSession(omg.Session().GetId())
	msgStr := "MSG"
	future := concurrent.NewFuture()
	rs.OnMessage(func(msg *messages.SessionMessage) {
		if msg.GetMessage() == msgStr {
			future.Completable().Complete(nil)
		}
	})

	omg.Session().SendMessage(msgStr)
	assert.True(t, future.AwaitTimeout(Timeout*10).IsSuccess())
	assert.True(t, omg.Close().Await().IsSuccess())
}

func TestOmega_SelfSessionMeta(t *testing.T) {
	omg := NewOmegaBuilder(conf).Connect().Get()
	assert.NotNil(t, omg)
	rs := omg.GetRemoteSession(omg.Session().GetId())
	name := omg.Session().GetName()
	assert.True(t, omg.Session().SetMetadata(map[string]any{"KEY": "VALUE"}).AwaitTimeout(Timeout*10).IsSuccess())
	assert.Equal(t, name, omg.Session().GetName())
	assert.Nil(t, rs.GetMetadata()["KEY"])
	assert.True(t, rs.Fetch().AwaitTimeout(Timeout*10).IsSuccess())
	assert.Equal(t, "VALUE", rs.GetMetadata()["KEY"])
	assert.Equal(t, name, rs.GetName())
	assert.True(t, omg.Close().Await().IsSuccess())
}

func TestOmega_SessionClose(t *testing.T) {
	omg := NewOmegaBuilder(conf).Connect().Get()
	assert.NotNil(t, omg)
	rs := omg.GetRemoteSession(omg.Session().GetId())
	assert.True(t, rs.Base().Close().AwaitTimeout(Timeout*10).IsSuccess())
	assert.False(t, omg.Session().Base().IsClosed())
	assert.True(t, omg.Close().Await().IsSuccess())
}
